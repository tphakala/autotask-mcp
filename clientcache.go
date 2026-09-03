package main

import (
	"container/list"
	"crypto/sha256"
	"errors"
	"sync"

	"github.com/tphakala/autotask-mcp/services"
	autotask "github.com/tphakala/go-autotask"
	"golang.org/x/sync/singleflight"
)

// errCacheClosed is returned by getOrCreate once closeAll has run, so a request that
// outlives the shutdown drain cannot insert a client the cache would never close.
var errCacheClosed = errors.New("client cache is closed")

// tenantClient bundles a per-tenant autotask client with its warm metadata caches.
// Reusing the same bundle across sessions from one tenant keeps the MappingCache and
// PicklistCache warm instead of rebuilding them cold on every reconnect.
type tenantClient struct {
	client   *autotask.Client
	mapper   *services.MappingCache
	picklist *services.PicklistCache
}

// closeTenant closes a tenant's underlying client. In gateway configuration this is a
// no-op today (the client registers no closers: the threshold monitor, its only closer, is
// deliberately omitted, and the rate limiter and circuit breaker register none), but it is
// called at shutdown so the behavior stays correct if go-autotask ever registers a real
// closer. Note the rate limiter and circuit breaker are not stateless: their tokens and
// breaker state are now shared across a tenant's concurrent sessions.
func closeTenant(tc *tenantClient) {
	if tc == nil || tc.client == nil {
		return
	}
	_ = tc.client.Close()
}

// credentialKey derives a stable, non-reversible cache key from a tenant's Autotask
// credentials. The 0x00 delimiters between fields prevent boundary-shift collisions
// between adjacent fields (for example ("ab", "c", ...) and ("a", "bc", ...)). The raw
// digest bytes are used directly as the map key; the plaintext credentials are never
// stored or logged.
//
// SHA-256 is the right primitive here, not a password KDF (bcrypt/scrypt/argon2): this is
// an in-process cache key over credentials the process already holds in plaintext, not a
// stored password verifier. A deliberately slow, salted KDF would be non-deterministic
// (defeating the lookup) and would add latency to every request, while providing no benefit
// (the digest lives only in memory and never leaves the process, so there is nothing to
// brute-force offline). A static-analysis rule that flags "hashing a secret with SHA-256"
// as weak password hashing does not apply to this use.
func credentialKey(apiKey, apiSecret, integrationCode string) string {
	h := sha256.New()
	h.Write([]byte(apiKey))
	h.Write([]byte{0})
	h.Write([]byte(apiSecret))
	h.Write([]byte{0})
	h.Write([]byte(integrationCode))
	return string(h.Sum(nil))
}

// clientCache is an LRU cache of per-tenant autotask clients keyed by credentialKey. It
// bounds how many live clients gateway mode holds and warms each tenant's metadata caches
// across reconnects. It is safe for concurrent use.
//
// Eviction only drops the cache's reference to a client; it does NOT close it. Closing an
// evicted client would be a no-op today and would be unsafe the moment go-autotask gives
// Close real work, because a concurrent in-flight session may still hold that client.
// Instead the garbage collector reclaims an evicted client once the last request using it
// returns. Clients are closed only by closeAll at shutdown, and runHTTP waits for the
// server's Shutdown to finish draining in-flight connections before that deferred call
// runs, so under a normal drain no request still holds a client. As a backstop for a
// request that outlives the drain deadline, closeAll marks the cache closed and getOrCreate
// then refuses to insert, closing the freshly built client instead, so a client built
// during shutdown is never orphaned. Closing is a no-op today anyway (see closeTenant); if
// go-autotask ever registers a real closer for these clients, this drain-then-close
// ordering plus the closed guard is what keeps it safe, and deterministic mid-life cleanup
// would still need reference counting driven by a reliable session-end signal (the go-sdk
// v1.7.0 exposes no public per-session end callback usable from the getServer factory).
type clientCache struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List               // front = most recently used
	items    map[string]*list.Element // key -> element holding *cacheEntry
	group    singleflight.Group       // dedupes concurrent creation of the same key
	closed   bool                     // set by closeAll; blocks further insertion
	// closeFn closes an entry at shutdown. It is a field so tests can observe closes
	// with fake entries; production uses closeTenant.
	closeFn func(*tenantClient)
}

type cacheEntry struct {
	key    string
	tenant *tenantClient
}

// newClientCache creates an LRU cache bounded to capacity entries. A capacity below 1 is
// clamped to 1 so the cache never degenerates into one that evicts every insertion.
func newClientCache(capacity int) *clientCache {
	if capacity < 1 {
		capacity = 1
	}
	return &clientCache{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[string]*list.Element),
		closeFn:  closeTenant,
	}
}

// getOrCreate returns the cached tenant for key, or calls create to build one and caches
// it. Concurrent calls for the same missing key are deduplicated so create runs exactly
// once (no redundant zone-discovery network I/O); create runs without the cache lock held,
// so different tenants are built in parallel. Inserting past capacity evicts the
// least-recently-used entry (dropping the reference only; see the type doc).
func (c *clientCache) getOrCreate(key string, create func() (*tenantClient, error)) (*tenantClient, error) {
	// Fast path: an existing entry is promoted to most-recently-used and returned.
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errCacheClosed
	}
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		tenant := el.Value.(*cacheEntry).tenant
		c.mu.Unlock()
		return tenant, nil
	}
	c.mu.Unlock()

	// Miss: build (or wait for a concurrent build of) this key exactly once.
	v, err, _ := c.group.Do(key, func() (any, error) {
		// Re-check under the lock: another goroutine may have populated this key
		// between our fast-path miss and entering the singleflight group.
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			return nil, errCacheClosed
		}
		if el, ok := c.items[key]; ok {
			c.ll.MoveToFront(el)
			tenant := el.Value.(*cacheEntry).tenant
			c.mu.Unlock()
			return tenant, nil
		}
		c.mu.Unlock()

		// Build the client without holding the lock (network I/O). A failed build is
		// not cached: singleflight forgets the key when this returns, so the next
		// request retries.
		tenant, cerr := create()
		if cerr != nil {
			return nil, cerr
		}

		c.mu.Lock()
		if c.closed {
			// closeAll ran while we were building. Do not insert into a cleared cache
			// (closeAll's snapshot would never close this entry); discard the client.
			c.mu.Unlock()
			c.closeFn(tenant)
			return nil, errCacheClosed
		}
		el := c.ll.PushFront(&cacheEntry{key: key, tenant: tenant})
		c.items[key] = el
		if c.ll.Len() > c.capacity {
			if back := c.ll.Back(); back != nil {
				ev := back.Value.(*cacheEntry)
				c.ll.Remove(back)
				delete(c.items, ev.key)
				// Reference dropped, not closed; see the type doc. The entry just
				// pushed is at the front and capacity is at least 1, so it can
				// never be the one evicted here.
			}
		}
		c.mu.Unlock()
		return tenant, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*tenantClient), nil
}

// closeAll marks the cache closed, then drops and closes every cached entry. It is called
// once, on graceful shutdown, after runHTTP has waited for the server to drain in-flight
// connections. Marking the cache closed makes a getOrCreate that outlived the drain refuse
// to insert its freshly built client (getOrCreate closes it instead), so no client is
// orphaned by racing the teardown.
func (c *clientCache) closeAll() {
	c.mu.Lock()
	c.closed = true
	tenants := make([]*tenantClient, 0, len(c.items))
	for _, el := range c.items {
		tenants = append(tenants, el.Value.(*cacheEntry).tenant)
	}
	c.ll.Init()
	c.items = make(map[string]*list.Element)
	c.mu.Unlock()

	for _, tenant := range tenants {
		c.closeFn(tenant)
	}
}

// size reports the number of cached entries. Used by tests.
func (c *clientCache) size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.items)
}
