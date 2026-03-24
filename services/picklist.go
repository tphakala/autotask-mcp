package services

import (
	"context"
	"fmt"
	"sync"

	autotask "github.com/tphakala/go-autotask"
	"github.com/tphakala/go-autotask/metadata"
)

// PicklistCache is a lazy-loaded, indefinite cache for entity field and picklist metadata.
// Picklist values rarely change, so no TTL is applied.
type PicklistCache struct {
	client *autotask.Client
	mu     sync.RWMutex
	fields map[string][]metadata.FieldInfo // keyed by entity name
}

// NewPicklistCache creates a new PicklistCache using the provided Autotask client.
func NewPicklistCache(client *autotask.Client) *PicklistCache {
	return &PicklistCache{
		client: client,
		fields: make(map[string][]metadata.FieldInfo),
	}
}

// GetFields returns field metadata for the given entity name.
// Results are cached indefinitely after the first successful fetch.
func (p *PicklistCache) GetFields(ctx context.Context, entityName string) ([]metadata.FieldInfo, error) {
	// Check cache
	p.mu.RLock()
	if fields, ok := p.fields[entityName]; ok {
		p.mu.RUnlock()
		return fields, nil
	}
	p.mu.RUnlock()

	// Fetch from API
	fields, err := metadata.GetFields(ctx, p.client, entityName)
	if err != nil {
		return nil, fmt.Errorf("picklist: GetFields(%s): %w", entityName, err)
	}

	// Cache result
	p.mu.Lock()
	p.fields[entityName] = fields
	p.mu.Unlock()

	return fields, nil
}

// GetPicklistValues returns the picklist values for a specific field on an entity.
// Returns an error if the field is not found or is not a picklist field.
func (p *PicklistCache) GetPicklistValues(ctx context.Context, entityName, fieldName string) ([]metadata.PickListValue, error) {
	fields, err := p.GetFields(ctx, entityName)
	if err != nil {
		return nil, err
	}

	for _, f := range fields {
		if f.Name == fieldName {
			if !f.IsPickList {
				return nil, fmt.Errorf("picklist: field %q on %s is not a picklist field", fieldName, entityName)
			}
			return f.PickListValues, nil
		}
	}

	return nil, fmt.Errorf("picklist: field %q not found on %s", fieldName, entityName)
}
