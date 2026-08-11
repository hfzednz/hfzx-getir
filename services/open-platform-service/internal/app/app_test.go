package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/app"
	"github.com/nexora/open-platform-service/internal/app/memory"
	"github.com/nexora/open-platform-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Apps: r.Apps, Keys: r.Keys, Catalog: r.Catalog, Versions: r.Versions,
		Policies: r.Policies, Webhooks: r.Webhooks, Deliveries: r.Deliveries,
		SDKs: r.SDKs, Integrations: r.Integrations, Usage: r.Usage, Outbox: r.Outbox,
		HTTP: r.HTTP, Identity: r.Identity, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestOpenPlatformFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.SeedCatalog(ctx, tid); err != nil {
		t.Fatal(err)
	}
	cat, _ := d.ListCatalog(ctx, tid, "public")
	if len(cat) < 5 {
		t.Fatal(len(cat))
	}

	a, err := d.CreateApp(ctx, domain.DeveloperApp{
		TenantID: tid, Name: "Partner App", Kind: domain.AppPartner, OwnerEmail: "dev@example.test",
		Scopes: []string{"orders:read", "webhooks"},
	})
	if err != nil || a.OAuthClientID == "" {
		t.Fatalf("%+v %v", a, err)
	}

	k, err := d.CreateApiKey(ctx, tid, a.ID, "prod", []string{"orders:read"})
	if err != nil || k.SecretOnce == "" {
		t.Fatal(err)
	}

	_, err = d.UpsertPolicy(ctx, domain.GatewayPolicy{
		TenantID: tid, Name: "public-v1", RoutePrefix: "/v1/", TargetService: "bff-customer",
		RateLimitRPM: 1000, QuotaDaily: 100000, CanaryPercent: 10, RequireScopes: []string{"read"},
	})
	if err != nil {
		t.Fatal(err)
	}
	pols, _ := d.ExportGatewayPolicies(ctx, tid)
	if len(pols) != 1 {
		t.Fatal(pols)
	}

	wh, err := d.RegisterWebhook(ctx, domain.WebhookEndpoint{
		TenantID: tid, AppID: a.ID, URL: "https://partner.example/hooks",
		Events: []string{"order.*", "OrderCreated"},
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := d.EnqueueWebhook(ctx, tid, "OrderCreated", map[string]any{"orderId": "x"})
	if err != nil || n < 1 {
		t.Fatalf("%d %v", n, err)
	}
	delivered, err := d.ProcessDeliveries(ctx, 10)
	if err != nil || delivered < 1 {
		t.Fatalf("%d %v", delivered, err)
	}
	_ = wh

	sdk, err := d.PublishSdk(ctx, domain.SdkPackage{
		TenantID: tid, Language: domain.SdkGo, Name: "nexora-go", Version: "0.1.0",
	})
	if err != nil || sdk.Status != "published" {
		t.Fatal(err)
	}

	_, err = d.ConnectIntegration(ctx, domain.IntegrationConnector{
		TenantID: tid, AppID: a.ID, Kind: domain.IntERP, Provider: "sap",
	})
	if err != nil {
		t.Fatal(err)
	}

	_ = d.RecordUsage(ctx, tid, a.ID, true)
	st, _ := d.AdminStats(ctx, tid)
	if st["apps"].(int) < 1 || st["sdks"].(int) < 1 {
		t.Fatal(st)
	}
}
