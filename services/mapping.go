package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	autotask "github.com/tphakala/go-autotask"
)

const mappingTTL = 30 * time.Minute

type cacheEntry struct {
	name   string
	expiry time.Time
}

// MappingCache caches company and resource ID-to-name lookups with per-entry 30-minute TTL.
type MappingCache struct {
	client    *autotask.Client
	mu        sync.RWMutex
	companies map[int64]cacheEntry
	resources map[int64]cacheEntry
}

// NewMappingCache creates a new MappingCache using the provided Autotask client.
func NewMappingCache(client *autotask.Client) *MappingCache {
	return &MappingCache{
		client:    client,
		companies: make(map[int64]cacheEntry),
		resources: make(map[int64]cacheEntry),
	}
}

// GetCompanyName returns the company name for the given ID.
// Returns "" for id == 0. Returns "Unknown (ID)" on error.
func (m *MappingCache) GetCompanyName(ctx context.Context, id int64) string {
	if id == 0 {
		return ""
	}

	// Check cache first
	m.mu.RLock()
	if entry, ok := m.companies[id]; ok && time.Now().Before(entry.expiry) {
		m.mu.RUnlock()
		return entry.name
	}
	m.mu.RUnlock()

	// Fallback to API
	raw, err := autotask.GetRaw(ctx, m.client, "Companies", id)
	if err != nil {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	name, _ := raw["companyName"].(string)
	if name == "" {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	// Write to cache with per-entry TTL
	m.mu.Lock()
	m.companies[id] = cacheEntry{name: name, expiry: time.Now().Add(mappingTTL)}
	m.mu.Unlock()

	return name
}

// GetResourceName returns the full name (firstName + lastName) for the given resource ID.
// Returns "" for id == 0. Returns "Unknown (ID)" on error.
func (m *MappingCache) GetResourceName(ctx context.Context, id int64) string {
	if id == 0 {
		return ""
	}

	// Check cache first
	m.mu.RLock()
	if entry, ok := m.resources[id]; ok && time.Now().Before(entry.expiry) {
		m.mu.RUnlock()
		return entry.name
	}
	m.mu.RUnlock()

	// Fallback to API
	raw, err := autotask.GetRaw(ctx, m.client, "Resources", id)
	if err != nil {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	firstName, _ := raw["firstName"].(string)
	lastName, _ := raw["lastName"].(string)
	name := strings.TrimSpace(firstName + " " + lastName)
	if name == "" {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	// Write to cache with per-entry TTL
	m.mu.Lock()
	m.resources[id] = cacheEntry{name: name, expiry: time.Now().Add(mappingTTL)}
	m.mu.Unlock()

	return name
}

// EnhanceItems adds an "_enhanced" map with human-readable names to each item.
// It batch-preloads uncached company/resource names before enhancing to minimize API calls.
func (m *MappingCache) EnhanceItems(ctx context.Context, items []map[string]any) {
	// Collect unique IDs that need lookup.
	companyIDs := make(map[int64]bool)
	resourceIDs := make(map[int64]bool)
	for _, item := range items {
		if id, ok := toInt64(item["companyID"]); ok && id != 0 {
			companyIDs[id] = true
		}
		for _, field := range []string{"assignedResourceID", "resourceID", "projectLeadResourceID"} {
			if id, ok := toInt64(item[field]); ok && id != 0 {
				resourceIDs[id] = true
			}
		}
	}

	// Batch preload uncached entries.
	m.preloadCompanies(ctx, companyIDs)
	m.preloadResources(ctx, resourceIDs)

	// Now enhance using cached values (all lookups are cache hits).
	for _, item := range items {
		enhanced := make(map[string]any)

		if id, ok := toInt64(item["companyID"]); ok && id != 0 {
			enhanced["companyName"] = m.GetCompanyName(ctx, id)
		}
		if id, ok := toInt64(item["assignedResourceID"]); ok && id != 0 {
			enhanced["assignedResourceName"] = m.GetResourceName(ctx, id)
		}
		if id, ok := toInt64(item["resourceID"]); ok && id != 0 {
			enhanced["resourceName"] = m.GetResourceName(ctx, id)
		}
		if id, ok := toInt64(item["projectLeadResourceID"]); ok && id != 0 {
			enhanced["projectLeadResourceName"] = m.GetResourceName(ctx, id)
		}

		if len(enhanced) > 0 {
			item["_enhanced"] = enhanced
		}
	}
}

// preloadCompanies batch-fetches uncached company names via a single API query.
func (m *MappingCache) preloadCompanies(ctx context.Context, ids map[int64]bool) {
	uncached := m.filterUncached(ids, m.companies)
	if len(uncached) == 0 {
		return
	}

	q := autotask.NewQuery().
		Where("id", autotask.OpIn, uncached).
		Fields("id", "companyName").
		Limit(len(uncached))

	results, err := autotask.ListRaw(ctx, m.client, "Companies", q)
	if err != nil {
		return // fall back to per-ID lookups during enhance
	}

	now := time.Now()
	m.mu.Lock()
	for _, r := range results {
		id, ok := toInt64(r["id"])
		if !ok {
			continue
		}
		name, _ := r["companyName"].(string)
		if name != "" {
			m.companies[id] = cacheEntry{name: name, expiry: now.Add(mappingTTL)}
		}
	}
	m.mu.Unlock()
}

// preloadResources batch-fetches uncached resource names via a single API query.
func (m *MappingCache) preloadResources(ctx context.Context, ids map[int64]bool) {
	uncached := m.filterUncached(ids, m.resources)
	if len(uncached) == 0 {
		return
	}

	q := autotask.NewQuery().
		Where("id", autotask.OpIn, uncached).
		Fields("id", "firstName", "lastName").
		Limit(len(uncached))

	results, err := autotask.ListRaw(ctx, m.client, "Resources", q)
	if err != nil {
		return // fall back to per-ID lookups during enhance
	}

	now := time.Now()
	m.mu.Lock()
	for _, r := range results {
		id, ok := toInt64(r["id"])
		if !ok {
			continue
		}
		first, _ := r["firstName"].(string)
		last, _ := r["lastName"].(string)
		name := strings.TrimSpace(first + " " + last)
		if name != "" {
			m.resources[id] = cacheEntry{name: name, expiry: now.Add(mappingTTL)}
		}
	}
	m.mu.Unlock()
}

// filterUncached returns IDs from the set that are not in the cache (or expired).
func (m *MappingCache) filterUncached(ids map[int64]bool, cache map[int64]cacheEntry) []int64 {
	now := time.Now()
	m.mu.RLock()
	defer m.mu.RUnlock()

	var uncached []int64
	for id := range ids {
		entry, ok := cache[id]
		if !ok || now.After(entry.expiry) {
			uncached = append(uncached, id)
		}
	}
	return uncached
}

// toInt64 converts common numeric types to int64.
func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	}
	return 0, false
}
