package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app"
	"github.com/nexora/settlement-service/internal/app/memory"
	"github.com/nexora/settlement-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.LedgerClient, *memory.PayoutClient) {
	t.Helper()
	store := memory.NewStore()
	batches, events, outbox := memory.NewRepos(store)
	ledger := &memory.LedgerClient{}
	payout := &memory.PayoutClient{}
	deps := &app.Deps{
		Batches:   batches,
		Events:    events,
		Outbox:    outbox,
		Publisher: &memory.EventPublisher{S: store},
		Ledger:    ledger,
		Payout:    payout,
		Clock:     &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		IDs:       memory.IDGen{},
	}
	return deps, ledger, payout
}

func TestCreateApproveComplete(t *testing.T) {
	deps, ledger, payout := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()
	submitter := uuid.New()
	approver := uuid.New()

	b, err := deps.CreateBatch(ctx, app.CreateBatchInput{
		TenantID: tenant, Currency: "TRY",
		PeriodStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		ActorID:     submitter, IdempotencyKey: "batch-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err = deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenant, BatchID: b.ID, PayeeType: domain.PayeeCourier,
		PayeeRef: "courier-opaque-1", AmountMinor: 25000,
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err = deps.Submit(ctx, app.SubmitInput{TenantID: tenant, BatchID: b.ID, ActorID: submitter})
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != domain.BatchStatusPendingApproval {
		t.Fatalf("status=%s", b.Status)
	}
	b, err = deps.Approve(ctx, app.ApproveInput{TenantID: tenant, BatchID: b.ID, ActorID: approver})
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != domain.BatchStatusApproved {
		t.Fatalf("status=%s", b.Status)
	}
	b, err = deps.ExecutePayouts(ctx, app.ExecutePayoutsInput{TenantID: tenant, BatchID: b.ID, ActorID: approver})
	if err != nil {
		t.Fatal(err)
	}
	if b.Status != domain.BatchStatusCompleted {
		t.Fatalf("status=%s", b.Status)
	}
	if ledger.Calls != 1 || payout.Calls != 1 {
		t.Fatalf("ledger=%d payout=%d", ledger.Calls, payout.Calls)
	}
}

func TestRejectApproveBySameActorDualControl(t *testing.T) {
	deps, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()
	actor := uuid.New()

	b, err := deps.CreateBatch(ctx, app.CreateBatchInput{
		TenantID: tenant, Currency: "TRY", ActorID: actor,
		PeriodStart: time.Now().UTC(), PeriodEnd: time.Now().UTC(),
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenant, BatchID: b.ID, PayeeType: domain.PayeeMerchant,
		PayeeRef: "merchant-1", AmountMinor: 1000,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = deps.Submit(ctx, app.SubmitInput{TenantID: tenant, BatchID: b.ID, ActorID: actor})
	if err != nil {
		t.Fatal(err)
	}
	_, err = deps.Approve(ctx, app.ApproveInput{TenantID: tenant, BatchID: b.ID, ActorID: actor})
	if !errors.Is(err, domain.ErrDualControl) {
		t.Fatalf("expected ErrDualControl, got %v", err)
	}
}

func TestReconcileDetectsMismatch(t *testing.T) {
	deps, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()
	submitter := uuid.New()
	approver := uuid.New()

	b, _ := deps.CreateBatch(ctx, app.CreateBatchInput{
		TenantID: tenant, Currency: "TRY", ActorID: submitter,
		PeriodStart: time.Now().UTC(), PeriodEnd: time.Now().UTC(),
	})
	b, _ = deps.AddLine(ctx, app.AddLineInput{
		TenantID: tenant, BatchID: b.ID, PayeeType: domain.PayeePartner,
		PayeeRef: "partner-1", AmountMinor: 5000,
	})
	b, _ = deps.Submit(ctx, app.SubmitInput{TenantID: tenant, BatchID: b.ID, ActorID: submitter})
	b, _ = deps.Approve(ctx, app.ApproveInput{TenantID: tenant, BatchID: b.ID, ActorID: approver})
	b, err := deps.ExecutePayouts(ctx, app.ExecutePayoutsInput{TenantID: tenant, BatchID: b.ID, ActorID: approver})
	if err != nil {
		t.Fatal(err)
	}

	res, err := deps.ReconcileProviderReport(ctx, app.ReconcileProviderReportInput{
		TenantID: tenant, BatchID: b.ID, ProviderRef: "psp-report-1",
		ReportedMinor: 4800, ActorID: approver,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Matched {
		t.Fatal("expected mismatch")
	}
	if res.Mismatch == nil {
		t.Fatal("expected mismatch record")
	}
	if res.Mismatch.DeltaMinor != -200 {
		t.Fatalf("delta=%d", res.Mismatch.DeltaMinor)
	}
}

func TestCreateBatchIdempotent(t *testing.T) {
	deps, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.New()
	actor := uuid.New()
	in := app.CreateBatchInput{
		TenantID: tenant, Currency: "TRY",
		PeriodStart: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		ActorID:     actor, IdempotencyKey: "settle-dup-1",
	}
	a, err := deps.CreateBatch(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	b, err := deps.CreateBatch(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("duplicate settlement batches %s vs %s", a.ID, b.ID)
	}
}
