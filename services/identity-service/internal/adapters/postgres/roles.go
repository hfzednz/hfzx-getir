package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
)

type RoleRepo struct{ DB *sql.DB }

func (r *RoleRepo) GetRole(ctx context.Context, id uuid.UUID) (domain.Role, error) {
	var role domain.Role
	var tenant uuid.NullUUID
	var kind string
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, name, kind::text, description, created_at, updated_at FROM roles WHERE id=$1`, id).
		Scan(&role.ID, &tenant, &role.Name, &kind, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return domain.Role{}, mapNotFound(err)
	}
	role.TenantID = scanUUIDPtr(tenant)
	role.Kind = domain.RoleKind(kind)
	return role, nil
}

func (r *RoleRepo) GetRoleByName(ctx context.Context, tenantID *uuid.UUID, name string) (domain.Role, error) {
	var role domain.Role
	var tenant uuid.NullUUID
	var kind string
	var err error
	if tenantID == nil {
		err = r.DB.QueryRowContext(ctx, `
			SELECT id, tenant_id, name, kind::text, description, created_at, updated_at
			FROM roles WHERE tenant_id IS NULL AND name=$1`, name).
			Scan(&role.ID, &tenant, &role.Name, &kind, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	} else {
		err = r.DB.QueryRowContext(ctx, `
			SELECT id, tenant_id, name, kind::text, description, created_at, updated_at
			FROM roles WHERE tenant_id=$1 AND name=$2`, *tenantID, name).
			Scan(&role.ID, &tenant, &role.Name, &kind, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	}
	if err != nil {
		return domain.Role{}, mapNotFound(err)
	}
	role.TenantID = scanUUIDPtr(tenant)
	role.Kind = domain.RoleKind(kind)
	return role, nil
}

func (r *RoleRepo) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT p.id, p.resource, p.action, p.description, p.created_at
		FROM permissions p
		INNER JOIN role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id=$1`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Permission{}
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *RoleRepo) RoleGraph(ctx context.Context, tenantID uuid.UUID) (policy.RoleGraph, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, name, kind::text, description, created_at, updated_at
		FROM roles WHERE tenant_id IS NULL OR tenant_id=$1`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	graph := policy.RoleGraph{}
	for rows.Next() {
		var role domain.Role
		var tenant uuid.NullUUID
		var kind string
		if err := rows.Scan(&role.ID, &tenant, &role.Name, &kind, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		role.TenantID = scanUUIDPtr(tenant)
		role.Kind = domain.RoleKind(kind)
		perms, err := r.ListRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		parentRows, err := r.DB.QueryContext(ctx, `SELECT parent_role_id FROM role_parents WHERE role_id=$1`, role.ID)
		if err != nil {
			return nil, err
		}
		parents := []uuid.UUID{}
		func() {
			defer parentRows.Close()
			for parentRows.Next() {
				var pid uuid.UUID
				if scanErr := parentRows.Scan(&pid); scanErr != nil {
					err = scanErr
					return
				}
				parents = append(parents, pid)
			}
			if rowsErr := parentRows.Err(); rowsErr != nil {
				err = rowsErr
			}
		}()
		if err != nil {
			return nil, err
		}
		graph[role.ID] = policy.RoleNode{Role: role, ParentIDs: parents, Permissions: perms}
	}
	return graph, rows.Err()
}

func (r *RoleRepo) AssignRole(ctx context.Context, pr domain.PrincipalRole) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO principal_roles (id, principal_id, role_id, tenant_id, city_id, warehouse_id, granted_by, created_at, expires_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		pr.ID, pr.PrincipalID, pr.RoleID, nullUUID(pr.Scope.TenantID), nullUUID(pr.Scope.CityID), nullUUID(pr.Scope.WarehouseID),
		nullUUID(pr.GrantedBy), pr.CreatedAt, pr.ExpiresAt)
	return err
}

func (r *RoleRepo) ListPrincipalRoles(ctx context.Context, principalID uuid.UUID) ([]domain.PrincipalRole, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, role_id, tenant_id, city_id, warehouse_id, granted_by, created_at, expires_at
		FROM principal_roles WHERE principal_id=$1`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.PrincipalRole{}
	for rows.Next() {
		var pr domain.PrincipalRole
		var tenant, city, wh, granted uuid.NullUUID
		if err := rows.Scan(&pr.ID, &pr.PrincipalID, &pr.RoleID, &tenant, &city, &wh, &granted, &pr.CreatedAt, &pr.ExpiresAt); err != nil {
			return nil, err
		}
		pr.Scope = domain.Scope{TenantID: scanUUIDPtr(tenant), CityID: scanUUIDPtr(city), WarehouseID: scanUUIDPtr(wh)}
		pr.GrantedBy = scanUUIDPtr(granted)
		out = append(out, pr)
	}
	return out, rows.Err()
}

func (r *RoleRepo) CreateTemporaryGrant(ctx context.Context, g domain.TemporaryGrant) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO temporary_grants (id, principal_id, permission_id, tenant_id, city_id, warehouse_id, reason, granted_by, created_at, expires_at, revoked_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		g.ID, g.PrincipalID, g.PermissionID, nullUUID(g.Scope.TenantID), nullUUID(g.Scope.CityID), nullUUID(g.Scope.WarehouseID),
		g.Reason, nullUUID(g.GrantedBy), g.CreatedAt, g.ExpiresAt, g.RevokedAt)
	return err
}

func (r *RoleRepo) ListTemporaryGrants(ctx context.Context, principalID uuid.UUID) ([]domain.TemporaryGrant, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, principal_id, permission_id, tenant_id, city_id, warehouse_id, reason, granted_by, created_at, expires_at, revoked_at
		FROM temporary_grants WHERE principal_id=$1`, principalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.TemporaryGrant{}
	for rows.Next() {
		var g domain.TemporaryGrant
		var tenant, city, wh, granted uuid.NullUUID
		if err := rows.Scan(&g.ID, &g.PrincipalID, &g.PermissionID, &tenant, &city, &wh, &g.Reason, &granted, &g.CreatedAt, &g.ExpiresAt, &g.RevokedAt); err != nil {
			return nil, err
		}
		g.Scope = domain.Scope{TenantID: scanUUIDPtr(tenant), CityID: scanUUIDPtr(city), WarehouseID: scanUUIDPtr(wh)}
		g.GrantedBy = scanUUIDPtr(granted)
		out = append(out, g)
	}
	return out, rows.Err()
}

func (r *RoleRepo) GetPermission(ctx context.Context, id uuid.UUID) (domain.Permission, error) {
	var p domain.Permission
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, resource, action, description, created_at FROM permissions WHERE id=$1`, id).
		Scan(&p.ID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt)
	if err != nil {
		return domain.Permission{}, mapNotFound(err)
	}
	return p, nil
}

func (r *RoleRepo) FindPermission(ctx context.Context, resource, action string) (domain.Permission, error) {
	var p domain.Permission
	err := r.DB.QueryRowContext(ctx, `
		SELECT id, resource, action, description, created_at FROM permissions WHERE resource=$1 AND action=$2`, resource, action).
		Scan(&p.ID, &p.Resource, &p.Action, &p.Description, &p.CreatedAt)
	if err != nil {
		return domain.Permission{}, mapNotFound(err)
	}
	return p, nil
}

var _ ports.RoleRepository = (*RoleRepo)(nil)
