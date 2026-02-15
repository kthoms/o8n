// Package dao provides data access abstractions for Operaton resources
package dao

import "context"

// DAO defines the interface for fetching resources from Operaton
type DAO interface {
	// List fetches all resources of this type
	List(ctx context.Context) ([]interface{}, error)

	// Get fetches a specific resource by ID
	Get(ctx context.Context, id string) (interface{}, error)

	// Delete deletes a resource by ID (if supported)
	Delete(ctx context.Context, id string) error

	// Name returns the resource type name (e.g., "process-definitions")
	Name() string
}

// HierarchicalDAO extends DAO for resources that support drill-down navigation
type HierarchicalDAO interface {
	DAO

	// Children fetches child resources for a parent ID
	Children(ctx context.Context, parentID string) ([]interface{}, error)

	// ChildType returns the child resource type name
	ChildType() string
}

// ReadOnlyDAO extends DAO for resources that don't support deletion
type ReadOnlyDAO interface {
	// List fetches all resources of this type
	List(ctx context.Context) ([]interface{}, error)

	// Get fetches a specific resource by ID
	Get(ctx context.Context, id string) (interface{}, error)

	// Name returns the resource type name
	Name() string
}
