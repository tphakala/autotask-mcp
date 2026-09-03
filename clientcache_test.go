package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// closeRecorder records which tenants were closed, safely under concurrency.
type closeRecorder struct {
	mu     sync.Mutex
	closed []*tenantClient
}

func (r *closeRecorder) fn(tc *tenantClient) {
	r.mu.Lock()
	r.closed = append(r.closed, tc)
	r.mu.Unlock()
}

func (r *closeRecorder) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.closed)
}

func (r *closeRecorder) has(tc *tenantClient) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.closed {
		if c == tc {
			return true
		}
	}
	return false
}

// newTestCache returns a cache whose closes are recorded instead of touching a real client.
func newTestCache(capacity int) (*clientCache, *closeRecorder) {
	c := newClientCache(capacity)
	rec := &closeRecorder{}
	c.closeFn = rec.fn
	return c, rec
}

func makeTenant() (*tenantClient, error) { return &tenantClient{}, nil }

func TestCredentialKey(t *testing.T) {
	k := credentialKey("user", "secret", "int")
	if credentialKey("user", "secret", "int") != k {
		t.Fatal("same inputs must yield the same key")
	}
	if len(k) != sha256.Size {
		t.Fatalf("key length = %d, want %d", len(k), sha256.Size)
	}
	// Distinct credentials must map to distinct keys. Vary each field in isolation so
	// dropping any single field from the digest is caught (holding secret+integration
	// code constant while changing only the api key/username must still change the key).
	if credentialKey("userA", "s", "i") == credentialKey("userB", "s", "i") {
		t.Error("a differing api key/username must change the key")
	}
	if credentialKey("u", "secretA", "i") == credentialKey("u", "secretB", "i") {
		t.Error("a differing secret must change the key")
	}
	if credentialKey("a", "b", "c") == credentialKey("a", "b", "d") {
		t.Error("a differing integration code must change the key")
	}
	// The 0x00 delimiters must prevent boundary-shift collisions: without them,
	// ("ab","c",...) and ("a","bc",...) would hash the same concatenation.
	if credentialKey("ab", "c", "x") == credentialKey("a", "bc", "x") {
		t.Error("a field-boundary shift between key and secret must not collide")
	}
	if credentialKey("x", "ab", "c") == credentialKey("x", "a", "bc") {
		t.Error("a field-boundary shift between secret and integration code must not collide")
	}
}

func TestCloseTenant_NilSafe(t *testing.T) {
	// The production closer must not panic on a nil tenant or a tenant with no client;
	// the cache's fake closeFn never exercises these branches.
	closeTenant(nil)
	closeTenant(&tenantClient{})
}

func TestClientCache_GetOrCreateReuses(t *testing.T) {
	c, rec := newTestCache(4)
	first := &tenantClient{}
	creates := 0
	got1, err := c.getOrCreate("k", func() (*tenantClient, error) { creates++; return first, nil })
	if err != nil || got1 != first {
		t.Fatalf("first getOrCreate: got %p err %v", got1, err)
	}
	got2, err := c.getOrCreate("k", func() (*tenantClient, error) { creates++; return &tenantClient{}, nil })
	if err != nil {
		t.Fatal(err)
	}
	if got2 != first {
		t.Error("a repeat key must return the cached tenant, not a new one")
	}
	if creates != 1 {
		t.Errorf("create called %d times, want 1 (second call must hit the cache)", creates)
	}
	if rec.count() != 0 {
		t.Errorf("no close expected on reuse, got %d", rec.count())
	}
}

func TestClientCache_EvictsLeastRecentlyUsed(t *testing.T) {
	c, rec := newTestCache(2)
	t1, _ := c.getOrCreate("k1", makeTenant)
	_, _ = c.getOrCreate("k2", makeTenant)
	// Touch k1 so k2 becomes the least-recently-used entry.
	if got, _ := c.getOrCreate("k1", makeTenant); got != t1 {
		t.Fatal("k1 should still be cached")
	}
	// Inserting k3 must evict k2 (the LRU), not k1.
	_, _ = c.getOrCreate("k3", makeTenant)
	if c.size() != 2 {
		t.Fatalf("size = %d, want 2 (bounded to capacity)", c.size())
	}
	// Eviction drops the reference; it must NOT close the client (a concurrent
	// in-flight session may still hold it).
	if rec.count() != 0 {
		t.Errorf("eviction must not close any client, closed %d", rec.count())
	}
	// k1 is still cached (no rebuild).
	if got, _ := c.getOrCreate("k1", func() (*tenantClient, error) {
		t.Error("k1 must not be rebuilt; it should still be cached")
		return &tenantClient{}, nil
	}); got != t1 {
		t.Error("k1 should return the original cached tenant")
	}
	// k2 was evicted, so the next access rebuilds it.
	rebuilt := false
	_, _ = c.getOrCreate("k2", func() (*tenantClient, error) { rebuilt = true; return &tenantClient{}, nil })
	if !rebuilt {
		t.Error("k2 was evicted and must be rebuilt on next access")
	}
}

func TestClientCache_CloseAll(t *testing.T) {
	c, rec := newTestCache(8)
	tenants := map[string]*tenantClient{}
	for _, k := range []string{"a", "b", "c"} {
		tc := &tenantClient{}
		tenants[k] = tc
		if _, err := c.getOrCreate(k, func() (*tenantClient, error) { return tc, nil }); err != nil {
			t.Fatal(err)
		}
	}
	c.closeAll()
	if c.size() != 0 {
		t.Errorf("size after closeAll = %d, want 0", c.size())
	}
	if rec.count() != 3 {
		t.Errorf("closeAll closed %d entries, want 3", rec.count())
	}
	for k, tc := range tenants {
		if !rec.has(tc) {
			t.Errorf("tenant %q was not closed by closeAll", k)
		}
	}
}

func TestClientCache_CreateErrorNotCached(t *testing.T) {
	c, rec := newTestCache(4)
	_, err := c.getOrCreate("k", func() (*tenantClient, error) { return nil, errors.New("boom") })
	if err == nil {
		t.Fatal("expected the create error to propagate")
	}
	if c.size() != 0 {
		t.Errorf("a failed create must not cache anything, size = %d", c.size())
	}
	// A later successful create for the same key must work (singleflight forgets a
	// failed key, so it is not permanently poisoned).
	tc := &tenantClient{}
	got, err := c.getOrCreate("k", func() (*tenantClient, error) { return tc, nil })
	if err != nil || got != tc {
		t.Fatalf("retry after error: got %p err %v", got, err)
	}
	if rec.count() != 0 {
		t.Errorf("no close expected, got %d", rec.count())
	}
}

func TestClientCache_CapacityClampedToOne(t *testing.T) {
	c, _ := newTestCache(0)
	if c.capacity != 1 {
		t.Fatalf("capacity = %d, want 1 (a sub-1 capacity is clamped)", c.capacity)
	}
	_, _ = c.getOrCreate("a", makeTenant)
	_, _ = c.getOrCreate("b", makeTenant)
	if c.size() != 1 {
		t.Errorf("size = %d, want 1", c.size())
	}
}

// TestClientCache_ConcurrentSameKeyDedupes checks that a race to create the same missing
// key runs create exactly once (singleflight), that every caller observes that single
// result, and that nothing is closed. create blocks briefly so that without singleflight
// the goroutines would overlap and create many clients (the thundering herd this pins
// against occurs precisely when create does slow network I/O). Run with -race.
func TestClientCache_ConcurrentSameKeyDedupes(t *testing.T) {
	c, rec := newTestCache(4)
	const n = 64
	var creates atomic.Int32
	shared := &tenantClient{}
	results := make([]*tenantClient, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			tc, err := c.getOrCreate("shared", func() (*tenantClient, error) {
				creates.Add(1)
				time.Sleep(10 * time.Millisecond) // widen the race window
				return shared, nil
			})
			if err != nil {
				t.Errorf("getOrCreate: %v", err)
				return
			}
			results[i] = tc
		}(i)
	}
	wg.Wait()

	if got := creates.Load(); got != 1 {
		t.Errorf("create ran %d times, want 1 (singleflight must deduplicate)", got)
	}
	if c.size() != 1 {
		t.Fatalf("size = %d, want 1 (all callers share one key)", c.size())
	}
	for i, r := range results {
		if r != shared {
			t.Fatalf("result[%d] = %p, want the single shared tenant %p", i, r, shared)
		}
	}
	if rec.count() != 0 {
		t.Errorf("nothing should be closed, closed %d", rec.count())
	}
}

// TestClientCache_ConcurrentMixedKeys drives many goroutines over more keys than the
// capacity to exercise eviction under contention. Run with -race. The only invariant
// asserted here is that the cache never exceeds its bound.
func TestClientCache_ConcurrentMixedKeys(t *testing.T) {
	const capacity = 8
	c, _ := newTestCache(capacity)
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("k%d", i%20) // 20 keys vs capacity 8 forces evictions
			if _, err := c.getOrCreate(key, makeTenant); err != nil {
				t.Errorf("getOrCreate: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if c.size() > capacity {
		t.Errorf("size = %d exceeds capacity %d", c.size(), capacity)
	}
}
