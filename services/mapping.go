package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	autotask "github.com/tphakala/go-autotask"
)

const mappingTTL = 30 * time.Minute

// MappingCache caches company and resource ID-to-name lookups with a 30-minute TTL.
type MappingCache struct {
	client    *autotask.Client
	mu        sync.RWMutex
	companies map[int64]string
	resources map[int64]string
	expiry    time.Time
}

// NewMappingCache creates a new MappingCache using the provided Autotask client.
func NewMappingCache(client *autotask.Client) *MappingCache {
	return &MappingCache{
		client:    client,
		companies: make(map[int64]string),
		resources: make(map[int64]string),
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
	if time.Now().Before(m.expiry) {
		if name, ok := m.companies[id]; ok {
			m.mu.RUnlock()
			return name
		}
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

	// Write to cache
	m.mu.Lock()
	m.companies[id] = name
	m.expiry = time.Now().Add(mappingTTL)
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
	if time.Now().Before(m.expiry) {
		if name, ok := m.resources[id]; ok {
			m.mu.RUnlock()
			return name
		}
	}
	m.mu.RUnlock()

	// Fallback to API
	raw, err := autotask.GetRaw(ctx, m.client, "Resources", id)
	if err != nil {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	firstName, _ := raw["firstName"].(string)
	lastName, _ := raw["lastName"].(string)
	name := firstName + " " + lastName
	name = trimSpace(name)
	if name == "" {
		return fmt.Sprintf("Unknown (%d)", id)
	}

	// Write to cache
	m.mu.Lock()
	m.resources[id] = name
	m.expiry = time.Now().Add(mappingTTL)
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

// trimSpace trims leading/trailing whitespace from a string.
func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
