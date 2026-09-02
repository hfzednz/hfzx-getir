package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/app/memory"
	"github.com/nexora/platform-ops-service/internal/app/ports"
	"github.com/nexora/platform-ops-service/internal/domain"
)

func (d *Deps) registry() ports.Registry {
	if d.Registry == nil {
		d.Registry = memory.NewRegistry()
	}
	return d.Registry
}

func (d *Deps) ListTenants(ctx context.Context) (map[string]any, error) {
	items, props, err := d.registry().ListTenants(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"items": items, "total": len(items),
		"generatedAt": d.now(), "proposals": props,
	}, nil
}

func (d *Deps) GetTenant(ctx context.Context, id string) (map[string]any, error) {
	t, err := d.registry().GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	return tenantDetail(t), nil
}

func (d *Deps) CreateTenant(ctx context.Context, name, slug, companyID, isolation, region, actor string) (domain.PlatformTenant, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(slug) == "" {
		return domain.PlatformTenant{}, fmt.Errorf("%w: name and slug required", domain.ErrInvalidArgument)
	}
	if isolation == "" {
		isolation = "shared"
	}
	companyName := companyID
	if c, err := d.registry().GetCompany(ctx, companyID); err == nil {
		companyName = c.TradeName
		if companyName == "" {
			companyName = c.LegalName
		}
	}
	t := domain.PlatformTenant{
		ID: uuid.NewString(), Slug: slug, Name: name,
		CompanyID: companyID, CompanyName: companyName,
		IsolationMode: isolation, Status: "pending", Region: region,
		CreatedAt: d.now(), Description: name,
	}
	if err := d.registry().SaveTenant(ctx, t); err != nil {
		return domain.PlatformTenant{}, err
	}
	d.auditRegistry(ctx, actor, "tenants.create", "tenant", t.ID, nil, t.Name)
	return t, nil
}

func (d *Deps) PatchTenantIsolation(ctx context.Context, id, isolation, actor string) (map[string]any, error) {
	t, err := d.registry().GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	old := t.IsolationMode
	t.IsolationMode = isolation
	if err := d.registry().SaveTenant(ctx, t); err != nil {
		return nil, err
	}
	d.auditRegistry(ctx, actor, "tenants.isolation", "tenant", t.ID, &old, isolation)
	return tenantDetail(t), nil
}

func (d *Deps) ProposeTenantAction(ctx context.Context, tenantID, action, reason, requesterID, requesterEmail string) (domain.DualControlProposal, error) {
	t, err := d.registry().GetTenant(ctx, tenantID)
	if err != nil {
		return domain.DualControlProposal{}, err
	}
	if action != "tenant_suspend" && action != "tenant_delete" {
		return domain.DualControlProposal{}, fmt.Errorf("%w: unsupported action", domain.ErrInvalidArgument)
	}
	p := domain.DualControlProposal{
		ID: uuid.NewString(), Action: action, TenantID: tenantID, TenantName: t.Name,
		RequesterID: requesterID, RequesterEmail: requesterEmail,
		Status: "pending", Reason: reason, CreatedAt: d.now(),
	}
	if err := d.registry().SaveProposal(ctx, p); err != nil {
		return domain.DualControlProposal{}, err
	}
	d.auditRegistry(ctx, requesterEmail, "tenants.dual_control.propose", "tenant", tenantID, nil, action)
	return p, nil
}

func (d *Deps) ResolveTenantProposal(ctx context.Context, proposalID, decision, actor string) (domain.DualControlProposal, error) {
	p, err := d.registry().GetProposal(ctx, proposalID)
	if err != nil {
		return domain.DualControlProposal{}, err
	}
	if p.Status != "pending" {
		return domain.DualControlProposal{}, fmt.Errorf("%w: proposal already resolved", domain.ErrConflict)
	}
	if decision == "approved" {
		p.Status = "executed"
		t, err := d.registry().GetTenant(ctx, p.TenantID)
		if err == nil {
			if p.Action == "tenant_suspend" {
				t.Status = "suspended"
			} else if p.Action == "tenant_delete" {
				t.Status = "suspended"
			}
			_ = d.registry().SaveTenant(ctx, t)
		}
	} else {
		p.Status = "rejected"
	}
	if err := d.registry().SaveProposal(ctx, p); err != nil {
		return domain.DualControlProposal{}, err
	}
	d.auditRegistry(ctx, actor, "tenants.dual_control.resolve", "proposal", proposalID, nil, decision)
	return p, nil
}

func (d *Deps) ListCompanies(ctx context.Context) (map[string]any, error) {
	items, err := d.registry().ListCompanies(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": items, "total": len(items), "generatedAt": d.now()}, nil
}

func (d *Deps) GetCompany(ctx context.Context, id string) (map[string]any, error) {
	c, err := d.registry().GetCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	return companyDetail(c), nil
}

func (d *Deps) CreateCompany(ctx context.Context, legal, trade, country, currency, actor string) (domain.PlatformCompany, error) {
	if strings.TrimSpace(legal) == "" || strings.TrimSpace(trade) == "" {
		return domain.PlatformCompany{}, fmt.Errorf("%w: legalName and tradeName required", domain.ErrInvalidArgument)
	}
	c := domain.PlatformCompany{
		ID: uuid.NewString(), LegalName: legal, TradeName: trade,
		CountryCode: country, Status: "draft", PrimaryCurrency: currency,
		CreatedAt: d.now(), DefaultLocale: "tr-TR", TimeZone: "Europe/Istanbul",
		PrimaryColor: "#0B6E6E",
	}
	if err := d.registry().SaveCompany(ctx, c); err != nil {
		return domain.PlatformCompany{}, err
	}
	d.auditRegistry(ctx, actor, "companies.create", "company", c.ID, nil, trade)
	return c, nil
}

func (d *Deps) PatchCompany(ctx context.Context, id string, patch map[string]any, actor string) (map[string]any, error) {
	c, err := d.registry().GetCompany(ctx, id)
	if err != nil {
		return nil, err
	}
	if v, ok := patch["legalName"].(string); ok && v != "" {
		c.LegalName = v
	}
	if v, ok := patch["tradeName"].(string); ok && v != "" {
		c.TradeName = v
	}
	if v, ok := patch["status"].(string); ok && v != "" {
		c.Status = v
	}
	if err := d.registry().SaveCompany(ctx, c); err != nil {
		return nil, err
	}
	d.auditRegistry(ctx, actor, "companies.update", "company", id, nil, c.TradeName)
	return companyDetail(c), nil
}

func (d *Deps) DeleteCompany(ctx context.Context, id, actor string) error {
	if err := d.registry().DeleteCompany(ctx, id); err != nil {
		return err
	}
	d.auditRegistry(ctx, actor, "companies.delete", "company", id, nil, "")
	return nil
}

func (d *Deps) RolesSnapshot(_ context.Context) map[string]any {
	roles := []map[string]any{}
	for _, name := range platformRoleNames() {
		roles = append(roles, map[string]any{
			"id": "role-" + name, "key": name, "label": roleLabel(name),
			"scope": roleScope(name), "members": 0, "inheritsFrom": nil,
			"description": "platform role " + name,
		})
	}
	return map[string]any{
		"generatedAt": d.now(), "roles": roles,
		"permissionTemplates": []map[string]any{
			{"id": "tmpl_read", "key": "tmpl_read_all", "label": "Read all platform", "permissions": []string{"dashboard:read", "tenants:read", "audit:read"}},
			{"id": "tmpl_tenant", "key": "tmpl_tenant_ops", "label": "Tenant operations", "permissions": []string{"tenants:read", "tenants:write", "tenants:suspend", "companies:write"}},
		},
		"approvalChains": []any{}, "inheritance": []any{}, "temporaryPermissions": []any{},
	}
}

func (d *Deps) OrgSnapshot(ctx context.Context) (map[string]any, error) {
	people, err := d.registry().ListPeople(ctx)
	if err != nil {
		return nil, err
	}
	if people == nil {
		people = []domain.PlatformPerson{}
	}
	return map[string]any{
		"generatedAt": d.now(),
		"organizations": []any{}, "departments": []any{}, "teams": []any{},
		"people": people,
	}, nil
}

func (d *Deps) AuditSnapshot(ctx context.Context, q string) (map[string]any, error) {
	items, err := d.registry().ListAudit(ctx, q)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.PlatformAuditEntry{}
	}
	return map[string]any{
		"generatedAt": d.now(), "total": len(items), "immutable": true, "items": items,
	}, nil
}

func (d *Deps) auditRegistry(ctx context.Context, actor, action, resource, resourceID string, old *string, newVal string) {
	nv := newVal
	_ = d.registry().AppendAudit(ctx, domain.PlatformAuditEntry{
		ID: uuid.NewString(), ActorID: actor, ActorEmail: actor,
		Action: action, Resource: resource, ResourceID: resourceID,
		When: d.now(), Where: "platform-ops", Device: "api",
		IP: "", SessionID: "", OldValue: old, NewValue: &nv,
		Severity: "info", Sealed: true,
	})
}

func tenantDetail(t domain.PlatformTenant) map[string]any {
	return map[string]any{
		"id": t.ID, "slug": t.Slug, "name": t.Name, "companyId": t.CompanyID,
		"companyName": t.CompanyName, "isolationMode": t.IsolationMode,
		"status": t.Status, "region": t.Region, "createdAt": t.CreatedAt,
		"description": t.Description,
		"config": map[string]any{
			"featurePack": "standard", "maxWarehouses": 50, "maxUsers": 200,
			"dataResidency": t.Region, "rlsEnabled": t.IsolationMode != "separate",
		},
		"customization": map[string]any{
			"primaryColor": "#0B6E6E", "logoUrl": "", "customDomain": nil, "whiteLabel": false,
		},
		"migration": map[string]any{
			"status": "idle", "targetMode": nil, "progressPct": 0,
			"lastMessage": "No migration scheduled", "updatedAt": time.Now().UTC(),
		},
		"backups": []any{}, "monitoring": []any{},
	}
}

func companyDetail(c domain.PlatformCompany) map[string]any {
	return map[string]any{
		"id": c.ID, "legalName": c.LegalName, "tradeName": c.TradeName,
		"countryCode": c.CountryCode, "status": c.Status, "tenantCount": c.TenantCount,
		"primaryCurrency": c.PrimaryCurrency, "createdAt": c.CreatedAt,
		"business": map[string]any{
			"industry": c.Industry, "taxId": c.TaxID, "vatNumber": c.VATNumber,
			"billingEmail": c.BillingEmail, "registeredAddress": c.RegisteredAddr,
		},
		"tax": map[string]any{
			"defaultTaxEngine": "tr-vat", "vatRegistered": c.VATNumber != "",
			"withholdingEnabled": false, "fiscalYearStartMonth": 1,
		},
		"locales": map[string]any{
			"defaultLocale": firstNonEmptyStr(c.DefaultLocale, "tr-TR"),
			"locales": []string{"tr-TR", "en-US"},
			"timeZone": firstNonEmptyStr(c.TimeZone, "Europe/Istanbul"),
			"currencies": []string{firstNonEmptyStr(c.PrimaryCurrency, "TRY")},
		},
		"domains": []any{},
		"branding": map[string]any{
			"primaryColor": firstNonEmptyStr(c.PrimaryColor, "#0B6E6E"),
			"secondaryColor": "#111827", "logoUrl": "", "faviconUrl": "",
		},
	}
}

func platformRoleNames() []string {
	return []string{
		"customer", "courier", "picker", "packer", "dispatcher", "city_ops",
		"support_agent", "finance_analyst", "admin", "super_admin",
		"service_account", "partner", "supplier", "merchant",
	}
}

func roleLabel(name string) string {
	return strings.ReplaceAll(name, "_", " ")
}

func roleScope(name string) string {
	switch name {
	case "super_admin", "service_account":
		return "global"
	case "admin", "city_ops", "finance_analyst", "support_agent":
		return "company"
	default:
		return "department"
	}
}

func firstNonEmptyStr(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
