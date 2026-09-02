package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/nexora/platform-ops-service/internal/adapters/postgres"
	"github.com/nexora/platform-ops-service/internal/app"
	"github.com/nexora/platform-ops-service/internal/domain"
)

func TestPostgresRegistryRestart(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := postgres.Open(dsn)
	if err != nil {
		t.Skip(err.Error())
	}
	t.Cleanup(func() { _ = db.Close() })
	for _, stmt := range []string{
		`CREATE TABLE IF NOT EXISTS platform_companies (
			id TEXT PRIMARY KEY, legal_name TEXT NOT NULL, trade_name TEXT NOT NULL,
			country_code TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft',
			tenant_count INT NOT NULL DEFAULT 0, primary_currency TEXT NOT NULL DEFAULT 'TRY',
			industry TEXT NOT NULL DEFAULT '', tax_id TEXT NOT NULL DEFAULT '',
			vat_number TEXT NOT NULL DEFAULT '', billing_email TEXT NOT NULL DEFAULT '',
			registered_addr TEXT NOT NULL DEFAULT '', default_locale TEXT NOT NULL DEFAULT 'tr-TR',
			time_zone TEXT NOT NULL DEFAULT 'Europe/Istanbul', primary_color TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now())`,
		`CREATE TABLE IF NOT EXISTS platform_tenants (
			id TEXT PRIMARY KEY, slug TEXT NOT NULL UNIQUE, name TEXT NOT NULL,
			company_id TEXT NOT NULL DEFAULT '', company_name TEXT NOT NULL DEFAULT '',
			isolation_mode TEXT NOT NULL DEFAULT 'shared', status TEXT NOT NULL DEFAULT 'pending',
			region TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT now())`,
		`CREATE TABLE IF NOT EXISTS platform_dual_control (
			id TEXT PRIMARY KEY, action TEXT NOT NULL, tenant_id TEXT NOT NULL,
			tenant_name TEXT NOT NULL DEFAULT '', requester_id TEXT NOT NULL DEFAULT '',
			requester_email TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending',
			reason TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT now())`,
		`CREATE TABLE IF NOT EXISTS platform_registry_audit (
			id TEXT PRIMARY KEY, actor_id TEXT NOT NULL DEFAULT '', actor_email TEXT NOT NULL DEFAULT '',
			action TEXT NOT NULL, resource TEXT NOT NULL, resource_id TEXT NOT NULL DEFAULT '',
			occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(), loc TEXT NOT NULL DEFAULT '',
			device TEXT NOT NULL DEFAULT '', ip TEXT NOT NULL DEFAULT '', session_id TEXT NOT NULL DEFAULT '',
			old_value TEXT, new_value TEXT, severity TEXT NOT NULL DEFAULT 'info', sealed BOOLEAN NOT NULL DEFAULT TRUE)`,
		`CREATE TABLE IF NOT EXISTS platform_people (
			id TEXT PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL DEFAULT '',
			kind TEXT NOT NULL DEFAULT '', org_unit_id TEXT NOT NULL DEFAULT '',
			org_unit_name TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'active')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	reg1 := &postgres.Registry{DB: db}
	d := &app.Deps{Clock: app.SystemClock{}, IDs: app.UUIDGen{}, Registry: reg1}
	ctx := context.Background()
	co, err := d.CreateCompany(ctx, "Persist Co", "Persist", "TR", "TRY", "sa@nexora")
	if err != nil {
		t.Fatal(err)
	}
	ten, err := d.CreateTenant(ctx, "Persist QC", "persist-"+co.ID[:8], co.ID, "shared", "eu-west-1", "sa@nexora")
	if err != nil {
		t.Fatal(err)
	}
	reg2 := &postgres.Registry{DB: db}
	got, err := reg2.GetTenant(ctx, ten.ID)
	if err != nil || got.Name != "Persist QC" || got.IsolationMode != "shared" {
		t.Fatalf("restart get %+v %v", got, err)
	}
	if _, err := reg2.GetCompany(ctx, co.ID); err != nil {
		t.Fatal(err)
	}
	_ = domain.PlatformTenant{}
}
