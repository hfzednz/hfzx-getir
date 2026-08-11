package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Permission is a resource:action authorization atom.
type Permission struct {
	ID          uuid.UUID
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

// String returns the canonical "resource:action" form.
func (p Permission) String() string {
	return p.Resource + ":" + p.Action
}

// ParsePermission parses a "resource:action" string into parts.
func ParsePermission(s string) (resource, action string, err error) {
	s = strings.TrimSpace(s)
	resource, action, ok := strings.Cut(s, ":")
	if !ok || resource == "" || action == "" {
		return "", "", fmt.Errorf("%w: permission must be resource:action, got %q", ErrInvalidArgument, s)
	}
	if strings.Contains(action, ":") {
		return "", "", fmt.Errorf("%w: action must not contain ':'", ErrInvalidArgument)
	}
	return resource, action, nil
}

// NewPermission builds a Permission from resource and action.
func NewPermission(resource, action string) (Permission, error) {
	resource = strings.TrimSpace(resource)
	action = strings.TrimSpace(action)
	if resource == "" || action == "" {
		return Permission{}, fmt.Errorf("%w: resource and action required", ErrInvalidArgument)
	}
	if strings.Contains(resource, ":") || strings.Contains(action, ":") {
		return Permission{}, fmt.Errorf("%w: resource/action must not contain ':'", ErrInvalidArgument)
	}
	return Permission{Resource: resource, Action: action}, nil
}

func (p Permission) Validate() error {
	if p.Resource == "" || p.Action == "" {
		return fmt.Errorf("%w: resource and action required", ErrInvalidArgument)
	}
	if strings.Contains(p.Resource, ":") || strings.Contains(p.Action, ":") {
		return fmt.Errorf("%w: resource/action must not contain ':'", ErrInvalidArgument)
	}
	return nil
}

// Matches reports equality in resource:action form (supports "*" wildcards on either side).
func (p Permission) Matches(required Permission) bool {
	resOK := p.Resource == "*" || p.Resource == required.Resource
	actOK := p.Action == "*" || p.Action == required.Action
	return resOK && actOK
}
