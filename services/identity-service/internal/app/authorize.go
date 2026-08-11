package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
)

// CheckPermissionInput is a PDP authorization request.
type CheckPermissionInput struct {
	PrincipalID uuid.UUID
	Permission  string // resource:action
	Attrs       policy.Attributes
	Requirement policy.Requirement
}

// CheckPermissionResult is the PDP decision.
type CheckPermissionResult struct {
	Allow   bool
	Reasons []string
}

// CheckPermission expands RBAC (with inheritance) and applies ABAC requirements.
func (d *Deps) CheckPermission(ctx context.Context, in CheckPermissionInput) (CheckPermissionResult, error) {
	if in.PrincipalID == uuid.Nil || in.Permission == "" {
		return CheckPermissionResult{}, fmt.Errorf("%w: principal_id and permission required", domain.ErrInvalidArgument)
	}
	resource, action, err := domain.ParsePermission(in.Permission)
	if err != nil {
		return CheckPermissionResult{}, err
	}
	required, err := domain.NewPermission(resource, action)
	if err != nil {
		return CheckPermissionResult{}, err
	}

	p, err := d.Principals.GetByID(ctx, in.PrincipalID)
	if err != nil {
		return CheckPermissionResult{}, err
	}
	if !p.IsActive() {
		return CheckPermissionResult{Allow: false, Reasons: []string{"principal not active"}}, nil
	}

	effective, err := d.ListEffectivePermissions(ctx, in.PrincipalID)
	if err != nil {
		return CheckPermissionResult{}, err
	}
	if !policy.HasPermission(effective, required) {
		d.appendAudit(ctx, domain.AuditEvent{
			TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
			Action: "authz.deny", ResourceType: "permission", ResourceID: in.Permission,
			Outcome: domain.AuditOutcomeDenied,
		})
		return CheckPermissionResult{Allow: false, Reasons: []string{"permission not granted"}}, nil
	}

	// Default MaxRisk to 100 when unset (zero value would deny all).
	req := in.Requirement
	if req.MaxRisk == 0 && !req.RequireTrustedDevice && req.MinMFALevel == 0 &&
		req.RequiredTenant == nil && req.RequiredCity == nil && req.RequiredWarehouse == nil {
		req.MaxRisk = 100
	}
	if req.MaxRisk == 0 {
		req.MaxRisk = 100
	}
	attrs := in.Attrs
	if attrs.TenantID == nil {
		tid := p.TenantID
		attrs.TenantID = &tid
	}
	dec := policy.Evaluate(attrs, req)
	if !dec.Allow {
		d.appendAudit(ctx, domain.AuditEvent{
			TenantID: &p.TenantID, ActorID: &p.ID, ActorKind: string(p.Kind),
			Action: "authz.deny", ResourceType: "permission", ResourceID: in.Permission,
			Outcome: domain.AuditOutcomeDenied, Details: map[string]any{"reasons": dec.Reasons},
		})
		return CheckPermissionResult{Allow: false, Reasons: dec.Reasons}, nil
	}
	return CheckPermissionResult{Allow: true}, nil
}

// ListEffectivePermissions returns merged RBAC + temporary grant permissions.
func (d *Deps) ListEffectivePermissions(ctx context.Context, principalID uuid.UUID) ([]domain.Permission, error) {
	p, err := d.Principals.GetByID(ctx, principalID)
	if err != nil {
		return nil, err
	}
	bindings, err := d.Roles.ListPrincipalRoles(ctx, principalID)
	if err != nil {
		return nil, err
	}
	now := d.now()
	roleIDs := make([]uuid.UUID, 0, len(bindings))
	for _, b := range bindings {
		if b.IsActive(now) {
			roleIDs = append(roleIDs, b.RoleID)
		}
	}
	graph, err := d.Roles.RoleGraph(ctx, p.TenantID)
	if err != nil {
		return nil, err
	}

	grants, err := d.Roles.ListTemporaryGrants(ctx, principalID)
	if err != nil {
		return nil, err
	}
	tempPerms := make([]domain.Permission, 0)
	for _, g := range grants {
		if !g.IsActive(now) {
			continue
		}
		perm, err := d.Roles.GetPermission(ctx, g.PermissionID)
		if err != nil {
			continue
		}
		tempPerms = append(tempPerms, perm)
	}
	return policy.EffectivePermissions(graph, roleIDs, tempPerms)
}
