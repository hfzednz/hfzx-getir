package policy

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
)

// RoleNode is a role plus its direct parent IDs and attached permissions.
type RoleNode struct {
	Role        domain.Role
	ParentIDs   []uuid.UUID
	Permissions []domain.Permission
}

// RoleGraph holds roles keyed by ID for inheritance expansion.
type RoleGraph map[uuid.UUID]RoleNode

// ExpandRoles walks role inheritance (parents) and returns the effective permission set.
// Inheritance is parent → child: a role inherits all permissions of its ancestors.
// Cycles are detected and returned as an error.
func ExpandRoles(graph RoleGraph, roleIDs []uuid.UUID) ([]domain.Permission, error) {
	seenPerms := make(map[string]domain.Permission)
	visiting := make(map[uuid.UUID]bool)
	visited := make(map[uuid.UUID]bool)

	var walk func(id uuid.UUID) error
	walk = func(id uuid.UUID) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("%w: role inheritance cycle at %s", domain.ErrInvariant, id)
		}
		node, ok := graph[id]
		if !ok {
			return fmt.Errorf("%w: role %s", domain.ErrNotFound, id)
		}
		visiting[id] = true
		for _, parentID := range node.ParentIDs {
			if err := walk(parentID); err != nil {
				return err
			}
		}
		visiting[id] = false
		visited[id] = true

		for _, p := range node.Permissions {
			seenPerms[p.String()] = p
		}
		return nil
	}

	for _, id := range roleIDs {
		if err := walk(id); err != nil {
			return nil, err
		}
	}

	out := make([]domain.Permission, 0, len(seenPerms))
	for _, p := range seenPerms {
		out = append(out, p)
	}
	return out, nil
}

// HasPermission reports whether the effective set grants the required permission.
func HasPermission(effective []domain.Permission, required domain.Permission) bool {
	for _, p := range effective {
		if p.Matches(required) {
			return true
		}
	}
	return false
}

// MergePermissions unions permission slices by resource:action key.
func MergePermissions(sets ...[]domain.Permission) []domain.Permission {
	seen := make(map[string]domain.Permission)
	for _, set := range sets {
		for _, p := range set {
			seen[p.String()] = p
		}
	}
	out := make([]domain.Permission, 0, len(seen))
	for _, p := range seen {
		out = append(out, p)
	}
	return out
}

// EffectivePermissions expands assigned roles and merges temporary grants that are active.
func EffectivePermissions(
	graph RoleGraph,
	roleIDs []uuid.UUID,
	temporary []domain.Permission,
) ([]domain.Permission, error) {
	expanded, err := ExpandRoles(graph, roleIDs)
	if err != nil {
		return nil, err
	}
	return MergePermissions(expanded, temporary), nil
}
