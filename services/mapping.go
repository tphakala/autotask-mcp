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

// EnhanceItems iterates items and adds an "_enhanced" map with human-readable names.
// Looks up companyID, assignedResourceID, resourceID, and projectLeadResourceID.
func (m *MappingCache) EnhanceItems(ctx context.Context, items []map[string]any) {
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
