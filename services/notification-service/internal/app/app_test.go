package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app"
	"github.com/nexora/notification-service/internal/app/memory"
	"github.com/nexora/notification-service/internal/domain"
)

type testEnv struct {
	Deps   *app.Deps
	Store  *memory.Store
	Repos  *memory.Repos
	Clock  *memory.Clock
	Email  *memory.MockEmail
	Push   *memory.MockPush
	SMS    *memory.MockSMS
	WA     *memory.MockWhatsApp
	Tenant uuid.UUID
	User   uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	email := &memory.MockEmail{}
	push := &memory.MockPush{}
	sms := &memory.MockSMS{}
	wa := &memory.MockWhatsApp{}
	deps := &app.Deps{
		Templates: repos.Templates, Messages: repos.Messages, Preferences: repos.Preferences,
		Devices: repos.Devices, Inbox: repos.Inbox, Schedules: repos.Schedules,
		Deliveries: repos.Deliveries, Outbox: repos.Outbox,
		Push: push, Email: email, SMS: sms, WhatsApp: wa,
		Publisher: &memory.EventPublisher{S: store},
		Clock:     clock,
		IDs:       memory.IDGen{},
	}
	return &testEnv{
		Deps: deps, Store: store, Repos: repos, Clock: clock,
		Email: email, Push: push, SMS: sms, WA: wa,
		Tenant: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		User:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}

func seedTemplate(t *testing.T, env *testEnv, key string, channel domain.Channel, subject, body string) {
	t.Helper()
	_, err := env.Deps.UpsertTemplate(context.Background(), app.UpsertTemplateInput{
		TenantID: env.Tenant, Key: key, Channel: channel, Locale: "en",
		Subject: subject, Body: body, Activate: true,
	})
	if err != nil {
		t.Fatalf("seed template: %v", err)
	}
}

func TestSendEmailTransactionalSucceeds(t *testing.T) {
	env := testDeps(t)
	seedTemplate(t, env, "welcome", domain.ChannelEmail, "Hi", "Hello {{name}}")
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, TemplateKey: "welcome",
		Recipient: "a@example.com", Vars: map[string]string{"name": "Ada"},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if msg.Status != domain.MessageSent {
		t.Fatalf("status=%s want sent", msg.Status)
	}
	if len(env.Email.Calls) != 1 {
		t.Fatalf("email calls=%d", len(env.Email.Calls))
	}
}

func TestMarketingSuppressedWhenOptedOut(t *testing.T) {
	env := testDeps(t)
	_, err := env.Deps.SetPreferences(context.Background(), app.SetPreferencesInput{
		TenantID: env.Tenant, PrincipalID: env.User,
		ChannelOptOut: map[string]bool{"email": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityMarketing, Subject: "Sale", Body: "Buy",
		Recipient: "a@example.com",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if msg.Status != domain.MessageSuppressed || msg.SuppressReason != "opt_out" {
		t.Fatalf("got status=%s reason=%s", msg.Status, msg.SuppressReason)
	}
	if len(env.Email.Calls) != 0 {
		t.Fatal("provider should not be called")
	}
}

func TestTransactionalSendsWhenMarketingOptedOut(t *testing.T) {
	env := testDeps(t)
	_, err := env.Deps.SetPreferences(context.Background(), app.SetPreferencesInput{
		TenantID: env.Tenant, PrincipalID: env.User,
		ChannelOptOut: map[string]bool{"email": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, Subject: "Receipt", Body: "Thanks",
		Recipient: "a@example.com",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if msg.Status != domain.MessageSent {
		t.Fatalf("status=%s", msg.Status)
	}
	if len(env.Email.Calls) != 1 {
		t.Fatal("expected email send")
	}
}

func TestQuietHoursSuppressMarketing(t *testing.T) {
	env := testDeps(t)
	env.Clock.T = time.Date(2026, 8, 6, 23, 30, 0, 0, time.UTC)
	qs, qe := 22, 7
	_, err := env.Deps.SetPreferences(context.Background(), app.SetPreferencesInput{
		TenantID: env.Tenant, PrincipalID: env.User, QuietStart: &qs, QuietEnd: &qe,
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityMarketing, Subject: "Night", Body: "Promo",
		Recipient: "a@example.com",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	if msg.Status != domain.MessageSuppressed || msg.SuppressReason != "quiet_hours" {
		t.Fatalf("got status=%s reason=%s", msg.Status, msg.SuppressReason)
	}
}

func TestTemplateNameRendered(t *testing.T) {
	env := testDeps(t)
	seedTemplate(t, env, "greet", domain.ChannelEmail, "Hello {{name}}", "Body {{name}}")
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, TemplateKey: "greet",
		Recipient: "a@example.com", Vars: map[string]string{"name": "Nexora"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Subject != "Hello Nexora" || msg.Body != "Body Nexora" {
		t.Fatalf("subject=%q body=%q", msg.Subject, msg.Body)
	}
	if env.Email.Calls[0].Subject != "Hello Nexora" {
		t.Fatalf("provider subject=%q", env.Email.Calls[0].Subject)
	}
}

func TestPushUsesDeviceToken(t *testing.T) {
	env := testDeps(t)
	_, err := env.Deps.RegisterDevice(context.Background(), app.RegisterDeviceInput{
		TenantID: env.Tenant, PrincipalID: env.User,
		Platform: domain.PlatformFCM, Token: "device-token-abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelPush,
		Priority: domain.PriorityTransactional, Subject: "Hi", Body: "Push me",
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Status != domain.MessageSent {
		t.Fatalf("status=%s", msg.Status)
	}
	if len(env.Push.Calls) != 1 || env.Push.Calls[0].Token != "device-token-abc" {
		t.Fatalf("push calls=%+v", env.Push.Calls)
	}
}

func TestScheduleThenProcessDueSends(t *testing.T) {
	env := testDeps(t)
	sendAt := env.Clock.T.Add(1 * time.Hour)
	sch, err := env.Deps.ScheduleSend(context.Background(), app.ScheduleSendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, Subject: "Later", Body: "Soon",
		Recipient: "a@example.com", SendAt: sendAt,
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := env.Deps.ProcessDueSchedules(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected 0 due, got %d", n)
	}
	env.Clock.Advance(2 * time.Hour)
	n, err = env.Deps.ProcessDueSchedules(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("expected 1 processed, got %d", n)
	}
	got, err := env.Repos.Schedules.Get(context.Background(), env.Tenant, sch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.ScheduleProcessed || got.MessageID == nil {
		t.Fatalf("schedule=%+v", got)
	}
	if len(env.Email.Calls) != 1 {
		t.Fatal("expected email after process")
	}
}

func TestProviderFailRetryThenDLQ(t *testing.T) {
	env := testDeps(t)
	env.Email.Fail = errors.New("smtp down")
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, Subject: "X", Body: "Y",
		Recipient: "a@example.com",
	})
	if !errors.Is(err, domain.ErrProviderFailed) {
		t.Fatalf("err=%v", err)
	}
	if msg.Status != domain.MessageFailed || msg.Attempts != 1 {
		t.Fatalf("msg=%+v", msg)
	}
	// attempt 2
	msg, err = env.Deps.RetryFailed(context.Background(), app.RetryFailedInput{
		TenantID: env.Tenant, MessageID: msg.ID,
	})
	if !errors.Is(err, domain.ErrProviderFailed) || msg.Attempts != 2 {
		t.Fatalf("retry1 attempts=%d err=%v", msg.Attempts, err)
	}
	// attempt 3 → DLQ
	msg, err = env.Deps.RetryFailed(context.Background(), app.RetryFailedInput{
		TenantID: env.Tenant, MessageID: msg.ID,
	})
	if !errors.Is(err, domain.ErrProviderFailed) || msg.Attempts != 3 {
		t.Fatalf("retry2 attempts=%d err=%v", msg.Attempts, err)
	}
	dlq, err := env.Repos.Deliveries.ListDLQ(context.Background(), env.Tenant, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(dlq) == 0 {
		t.Fatal("expected DLQ entry")
	}
}

func TestInboxItemCreatedForInApp(t *testing.T) {
	env := testDeps(t)
	msg, err := env.Deps.Send(context.Background(), app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelInApp,
		Priority: domain.PrioritySystem, Subject: "Title", Body: "In-app body",
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.Status != domain.MessageSent {
		t.Fatalf("status=%s", msg.Status)
	}
	items, err := env.Deps.ListInbox(context.Background(), env.Tenant, env.User, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Body != "In-app body" {
		t.Fatalf("inbox=%+v", items)
	}
}

func TestIdempotentSendSameKey(t *testing.T) {
	env := testDeps(t)
	in := app.SendInput{
		TenantID: env.Tenant, PrincipalID: env.User, Channel: domain.ChannelEmail,
		Priority: domain.PriorityTransactional, Subject: "Once", Body: "Only",
		Recipient: "a@example.com", IdempotencyKey: "idem-1",
	}
	m1, err := env.Deps.Send(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := env.Deps.Send(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if m1.ID != m2.ID {
		t.Fatalf("ids differ %s vs %s", m1.ID, m2.ID)
	}
	if len(env.Email.Calls) != 1 {
		t.Fatalf("expected 1 provider call, got %d", len(env.Email.Calls))
	}
}
