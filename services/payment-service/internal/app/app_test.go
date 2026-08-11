package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/adapters/psp"
	"github.com/nexora/payment-service/internal/app"
	"github.com/nexora/payment-service/internal/app/memory"
	"github.com/nexora/payment-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Store     *memory.Store
	Primary   *psp.MockPSP
	Failover  *psp.MockPSP
	Router    *psp.Failover
	Fraud     *memory.FraudClient
	Tenant    uuid.UUID
	Principal uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	intents, outbox := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	primary := psp.NewMock("mock_primary")
	secondary := psp.NewMock("mock_failover")
	router := psp.NewFailover(primary, secondary)
	fraud := &memory.FraudClient{RiskScore: 10, Decision: domain.FraudAllow}
	deps := &app.Deps{
		Intents: intents, Outbox: outbox,
		Publisher: &memory.EventPublisher{S: store},
		PSP: router, Fraud: fraud,
		Wallet: &memory.WalletClient{Success: true},
		Ledger: &memory.LedgerClient{},
		Clock:  clock, IDs: memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Store: store, Primary: primary, Failover: secondary, Router: router, Fraud: fraud,
		Tenant: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Principal: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}

func createAuth(t *testing.T, env *testEnv, idem string, amount int64) domain.PaymentIntent {
	t.Helper()
	ctx := context.Background()
	intent, err := env.Deps.CreateIntent(ctx, app.CreateIntentInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: "ord-opaque-1", AmountMinor: amount, Currency: "TRY",
		MethodType: domain.MethodCard, IdempotencyKey: idem,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	intent, err = env.Deps.Authorize(ctx, app.AuthorizeInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: idem + "-auth", Token: "tok_x",
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	return intent
}

func TestAuthorizeCapture(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	intent := createAuth(t, env, "ac-1", 5000)
	if intent.Status != domain.IntentAuthorized {
		t.Fatalf("status=%s", intent.Status)
	}
	intent, err := env.Deps.Capture(ctx, app.CaptureInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "ac-1-cap",
	})
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if intent.Status != domain.IntentCaptured || intent.CapturedMinor != 5000 {
		t.Fatalf("got status=%s captured=%d", intent.Status, intent.CapturedMinor)
	}
}

func TestVoid(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	intent := createAuth(t, env, "void-1", 3000)
	intent, err := env.Deps.Void(ctx, app.VoidInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "void-1-k",
	})
	if err != nil {
		t.Fatalf("void: %v", err)
	}
	if intent.Status != domain.IntentVoided {
		t.Fatalf("status=%s", intent.Status)
	}
}

func TestRefundPartial(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	intent := createAuth(t, env, "ref-1", 10000)
	intent, err := env.Deps.Capture(ctx, app.CaptureInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "ref-1-cap",
	})
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	refund, intent, err := env.Deps.Refund(ctx, app.RefundInput{
		TenantID: env.Tenant, IntentID: intent.ID, AmountMinor: 2500,
		Reason: "partial", IdempotencyKey: "ref-1-r",
	})
	if err != nil {
		t.Fatalf("refund: %v", err)
	}
	if refund.Status != domain.RefundCompleted || refund.AmountMinor != 2500 {
		t.Fatalf("refund=%+v", refund)
	}
	if intent.RefundedMinor != 2500 || intent.Status != domain.IntentCaptured {
		t.Fatalf("intent refunded=%d status=%s", intent.RefundedMinor, intent.Status)
	}
}

func TestIdempotentAuthorize(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	intent, err := env.Deps.CreateIntent(ctx, app.CreateIntentInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: "o1", AmountMinor: 1000, Currency: "TRY",
		MethodType: domain.MethodCard, IdempotencyKey: "idem-auth",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	a1, err := env.Deps.Authorize(ctx, app.AuthorizeInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "idem-auth-a", Token: "tok",
	})
	if err != nil {
		t.Fatalf("auth1: %v", err)
	}
	a2, err := env.Deps.Authorize(ctx, app.AuthorizeInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "idem-auth-a", Token: "tok",
	})
	if err != nil {
		t.Fatalf("auth2: %v", err)
	}
	if a1.ID != a2.ID || a2.Status != domain.IntentAuthorized {
		t.Fatalf("idempotency failed")
	}
	if env.Primary.AuthCalls() != 1 {
		t.Fatalf("expected 1 auth call, got %d", env.Primary.AuthCalls())
	}
}

func TestFraudBlock(t *testing.T) {
	env := testDeps(t)
	env.Fraud.RiskScore = 95
	env.Fraud.Decision = domain.FraudBlock
	ctx := context.Background()
	intent, err := env.Deps.CreateIntent(ctx, app.CreateIntentInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: "o-fraud", AmountMinor: 99900, Currency: "TRY",
		MethodType: domain.MethodCard, IdempotencyKey: "fraud-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = env.Deps.Authorize(ctx, app.AuthorizeInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "fraud-1-a", Token: "tok",
	})
	if !errors.Is(err, domain.ErrFraudBlocked) {
		t.Fatalf("expected fraud block, got %v", err)
	}
	got, _ := env.Deps.GetIntent(ctx, env.Tenant, intent.ID)
	if got.Status != domain.IntentFailed {
		t.Fatalf("status=%s", got.Status)
	}
}

func TestFailoverWhenPrimaryFails(t *testing.T) {
	env := testDeps(t)
	env.Primary.SetFailAuthorize(true)
	ctx := context.Background()
	intent, err := env.Deps.CreateIntent(ctx, app.CreateIntentInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: "o-fo", AmountMinor: 2000, Currency: "TRY",
		MethodType: domain.MethodCard, IdempotencyKey: "fo-1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	intent, err = env.Deps.Authorize(ctx, app.AuthorizeInput{
		TenantID: env.Tenant, IntentID: intent.ID, IdempotencyKey: "fo-1-a", Token: "tok",
	})
	if err != nil {
		t.Fatalf("authorize: %v", err)
	}
	if intent.Status != domain.IntentAuthorized {
		t.Fatalf("status=%s", intent.Status)
	}
	if env.Router.LastUsed() != "mock_failover" {
		t.Fatalf("expected failover provider, got %s", env.Router.LastUsed())
	}
	if env.Primary.AuthCalls() != 1 || env.Failover.AuthCalls() != 1 {
		t.Fatalf("calls primary=%d failover=%d", env.Primary.AuthCalls(), env.Failover.AuthCalls())
	}
}
