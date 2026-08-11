package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app"
	"github.com/nexora/crm-service/internal/app/memory"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

type testEnv struct {
	Deps   *app.Deps
	Store  *memory.Store
	Repos  *memory.Repos
	Clock  *memory.Clock
	LLM    *memory.MockLLM
	Tenant uuid.UUID
	User   uuid.UUID
	Cust   uuid.UUID
	Agent  uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	llm := &memory.MockLLM{}
	deps := &app.Deps{
		Tickets: repos.Tickets, Chats: repos.Chats, Agents: repos.Agents,
		KB: repos.KB, Cases: repos.Cases, Feedback: repos.Feedback, SLA: repos.SLA,
		Outbox: repos.Outbox,
		Profiles: &memory.MockProfile{}, Orders: &memory.MockOrders{},
		Notify: &memory.MockNotify{}, Refunds: &memory.MockRefund{},
		LLM: llm, Publisher: &memory.EventPublisher{S: store},
		Clock: clock, IDs: memory.IDGen{}, AIConfidence: 0.5,
	}
	return &testEnv{
		Deps: deps, Store: store, Repos: repos, Clock: clock, LLM: llm,
		Tenant: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		User:   uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		Cust:   uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		Agent:  uuid.MustParse("44444444-4444-4444-4444-444444444444"),
	}
}

func TestCreateAssignResolveClose(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	ticket, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "Missing item",
		Priority: domain.PriorityNormal, Category: domain.CategoryOrder,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Status != domain.TicketStatusOpen {
		t.Fatalf("status=%s", ticket.Status)
	}
	ticket, err = env.Deps.AssignTicket(ctx, app.AssignTicketInput{
		TenantID: env.Tenant, TicketID: ticket.ID, AgentID: env.Agent,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Status != domain.TicketStatusInProgress {
		t.Fatalf("status=%s", ticket.Status)
	}
	ticket, err = env.Deps.Resolve(ctx, app.ResolveTicketInput{TenantID: env.Tenant, TicketID: ticket.ID})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Status != domain.TicketStatusResolved {
		t.Fatalf("status=%s", ticket.Status)
	}
	ticket, err = env.Deps.Close(ctx, app.CloseTicketInput{TenantID: env.Tenant, TicketID: ticket.ID})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Status != domain.TicketStatusClosed {
		t.Fatalf("status=%s", ticket.Status)
	}
}

func TestIllegalCloseFromOpen(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	ticket, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "Cannot close yet",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.Deps.Close(ctx, app.CloseTicketInput{TenantID: env.Tenant, TicketID: ticket.ID})
	if !errors.Is(err, domain.ErrIllegalTransition) {
		t.Fatalf("expected illegal transition, got %v", err)
	}
}

func TestEscalateRaisesPriority(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	ticket, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "Escalate me",
		Priority: domain.PriorityNormal,
	})
	if err != nil {
		t.Fatal(err)
	}
	ticket, esc, err := env.Deps.Escalate(ctx, app.EscalateTicketInput{
		TenantID: env.Tenant, TicketID: ticket.ID, Reason: "vip customer",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.Priority != domain.PriorityHigh {
		t.Fatalf("priority=%s", ticket.Priority)
	}
	if esc.FromPriority != domain.PriorityNormal || esc.ToPriority != domain.PriorityHigh {
		t.Fatalf("esc=%+v", esc)
	}
	escs, err := env.Repos.SLA.ListEscalations(ctx, env.Tenant, ticket.ID)
	if err != nil || len(escs) != 1 {
		t.Fatalf("escalations=%d err=%v", len(escs), err)
	}
}

func TestSLABreachDetected(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	_, err := env.Deps.UpsertSLAPolicy(ctx, app.UpsertSLAPolicyInput{
		TenantID: env.Tenant, Name: "normal", Priority: domain.PriorityNormal,
		FirstResponseMinutes: 30, ResolveMinutes: 60, Active: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ticket, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "SLA ticket",
		Priority: domain.PriorityNormal,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ticket.ResolveDue == nil {
		t.Fatal("expected resolve due")
	}
	env.Clock.Advance(2 * time.Hour)
	breached, err := env.Deps.EvaluateSLA(ctx, env.Tenant, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(breached) != 1 || !breached[0].SLABreached {
		t.Fatalf("breached=%+v", breached)
	}
	got, _ := env.Deps.Tickets.Get(ctx, env.Tenant, ticket.ID)
	if !got.SLABreached {
		t.Fatal("ticket not marked breached")
	}
}

func TestChatMessageCreatesConversation(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	conv, err := env.Deps.StartChat(ctx, app.StartChatInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Channel: "web",
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, c2, err := env.Deps.PostMessage(ctx, app.PostMessageInput{
		TenantID: env.Tenant, ConversationID: conv.ID,
		SenderRole: domain.SenderCustomer, Body: "Hello support",
	})
	if err != nil {
		t.Fatal(err)
	}
	if msg.ConversationID != conv.ID || c2.ID != conv.ID {
		t.Fatal("conversation mismatch")
	}
	msgs, err := env.Repos.Chats.ListMessages(ctx, env.Tenant, conv.ID)
	if err != nil || len(msgs) != 1 {
		t.Fatalf("msgs=%d err=%v", len(msgs), err)
	}
}

func TestAIAssistReplyAndLowConfidenceEscalate(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	art, err := env.Deps.UpsertArticle(ctx, app.UpsertArticleInput{
		TenantID: env.Tenant, Slug: "refunds", Title: "Refund policy",
		Body: "Refunds are processed within 3 days", Tags: []string{"refund"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.Deps.PublishArticle(ctx, env.Tenant, art.ID)
	if err != nil {
		t.Fatal(err)
	}

	env.LLM.ForceLowConf = false
	res, err := env.Deps.AIAssist(ctx, app.AIAssistInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Text: "How do refunds work?",
		AutoEscalate: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reply == "" {
		t.Fatal("expected reply")
	}
	if res.Escalated {
		t.Fatal("should not escalate on high confidence")
	}

	env.LLM.ForceLowConf = true
	res2, err := env.Deps.AIAssist(ctx, app.AIAssistInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Text: "Something confusing",
		AutoEscalate: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res2.Escalated || res2.Ticket == nil {
		t.Fatalf("expected escalation ticket, got %+v", res2)
	}
	if res2.Ticket.Priority != domain.PriorityUrgent && res2.Ticket.Priority != domain.PriorityHigh {
		// created high then escalated to urgent
		t.Fatalf("priority=%s", res2.Ticket.Priority)
	}
}

func TestKBSearchFindsArticle(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	art, err := env.Deps.UpsertArticle(ctx, app.UpsertArticleInput{
		TenantID: env.Tenant, Slug: "delivery-delays", Title: "Delivery delays",
		Body: "Weather may delay delivery", Tags: []string{"delivery"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = env.Deps.PublishArticle(ctx, env.Tenant, art.ID)
	if err != nil {
		t.Fatal(err)
	}
	hits, err := env.Deps.SearchKB(ctx, env.Tenant, "delivery")
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) == 0 {
		t.Fatal("expected KB hit")
	}
}

func TestCSATStored(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	csat, err := env.Deps.SubmitCSAT(ctx, app.SubmitCSATInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Score: 5, Comment: "Great",
	})
	if err != nil {
		t.Fatal(err)
	}
	if csat.Score != 5 {
		t.Fatalf("score=%d", csat.Score)
	}
	list, err := env.Repos.Feedback.ListCSAT(ctx, env.Tenant)
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%d err=%v", len(list), err)
	}
}

func TestCustomer360ReturnsStubs(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	env.Deps.Profiles = &memory.MockProfile{Profiles: map[uuid.UUID]ports.ProfileSummary{
		env.Cust: {CustomerID: env.Cust, DisplayName: "Hafize", Email: "h@example.com", Tier: "gold"},
	}}
	env.Deps.Orders = &memory.MockOrders{Orders: map[uuid.UUID][]ports.OrderSummary{
		env.Cust: {{
			OrderID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			Status: "delivered", Total: "99.00", Currency: "TRY",
			CreatedAt: env.Clock.T,
		}},
	}}
	_, _ = env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "360 ticket",
	})
	view, err := env.Deps.GetCustomer360(ctx, env.Tenant, env.Cust)
	if err != nil {
		t.Fatal(err)
	}
	if view.Profile.DisplayName != "Hafize" {
		t.Fatalf("profile=%+v", view.Profile)
	}
	if len(view.Orders) != 1 || view.Orders[0].Total != "99.00" {
		t.Fatalf("orders=%+v", view.Orders)
	}
	if len(view.Tickets) != 1 {
		t.Fatalf("tickets=%d", len(view.Tickets))
	}
}

func TestIdempotentCreateTicket(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	key := "idem-key-1"
	t1, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "First",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	t2, err := env.Deps.CreateTicket(ctx, app.CreateTicketInput{
		TenantID: env.Tenant, CustomerID: env.Cust, Subject: "Second ignored",
		IdempotencyKey: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	if t1.ID != t2.ID {
		t.Fatalf("expected same ticket id %s vs %s", t1.ID, t2.ID)
	}
	if t2.Subject != "First" {
		t.Fatalf("subject=%s", t2.Subject)
	}
}
