package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/app/ports"
	"github.com/nexora/crm-service/internal/domain"
)

// CreateCaseInput creates an investigation case.
type CreateCaseInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	TicketID   *uuid.UUID
	Type       string
	Title      string
	Details    string
	AssigneeID *uuid.UUID
}

// CreateCase creates a case; for refund type it may request refund via port (no execution).
func (d *Deps) CreateCase(ctx context.Context, in CreateCaseInput) (domain.Case, error) {
	if in.TenantID == uuid.Nil || in.CustomerID == uuid.Nil || strings.TrimSpace(in.Title) == "" {
		return domain.Case{}, fmt.Errorf("%w: tenant_id, customer_id, title required", domain.ErrInvalidArgument)
	}
	if !domain.ValidCaseType(in.Type) {
		return domain.Case{}, fmt.Errorf("%w: invalid case type", domain.ErrInvalidArgument)
	}
	now := d.now()
	c := domain.Case{
		ID: d.newID(), TenantID: in.TenantID, CustomerID: in.CustomerID,
		TicketID: in.TicketID, Type: in.Type, Status: domain.CaseStatusOpen,
		Title: strings.TrimSpace(in.Title), Details: in.Details, AssigneeID: in.AssigneeID,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Cases.Save(ctx, c); err != nil {
		return domain.Case{}, err
	}
	if in.Type == domain.CaseTypeComplaint {
		d.emit(ctx, c.TenantID, c.ID, domain.EventComplaintCreated, map[string]any{
			"customerId": c.CustomerID.String(), "title": c.Title,
		})
	}
	if in.Type == domain.CaseTypeRefund && d.Refunds != nil {
		reqID, _ := d.Refunds.RequestRefund(ctx, ports.RefundRequest{
			TenantID: in.TenantID, CustomerID: in.CustomerID,
			TicketID: uuid.Nil, Reason: in.Details,
		})
		d.emit(ctx, c.TenantID, c.ID, domain.EventRefundRequested, map[string]any{
			"requestId": reqID, "caseId": c.ID.String(),
		})
	}
	return c, nil
}

// UpdateCaseInput updates case fields.
type UpdateCaseInput struct {
	TenantID   uuid.UUID
	CaseID     uuid.UUID
	Status     string
	Title      string
	Details    string
	AssigneeID *uuid.UUID
}

// UpdateCase updates an existing case.
func (d *Deps) UpdateCase(ctx context.Context, in UpdateCaseInput) (domain.Case, error) {
	c, err := d.Cases.Get(ctx, in.TenantID, in.CaseID)
	if err != nil {
		return domain.Case{}, err
	}
	now := d.now()
	if in.Status != "" {
		c.Status = in.Status
	}
	if in.Title != "" {
		c.Title = in.Title
	}
	if in.Details != "" {
		c.Details = in.Details
	}
	if in.AssigneeID != nil {
		c.AssigneeID = in.AssigneeID
	}
	c.UpdatedAt = now
	if err := d.Cases.Save(ctx, c); err != nil {
		return domain.Case{}, err
	}
	return c, nil
}

// SubmitCSATInput records a CSAT score.
type SubmitCSATInput struct {
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	TicketID       *uuid.UUID
	ConversationID *uuid.UUID
	Score          int
	Comment        string
}

// SubmitCSAT stores a CSAT response (score 1–5).
func (d *Deps) SubmitCSAT(ctx context.Context, in SubmitCSATInput) (domain.CSATResponse, error) {
	if in.TenantID == uuid.Nil || in.CustomerID == uuid.Nil {
		return domain.CSATResponse{}, fmt.Errorf("%w: tenant_id, customer_id required", domain.ErrInvalidArgument)
	}
	if in.Score < 1 || in.Score > 5 {
		return domain.CSATResponse{}, fmt.Errorf("%w: csat score must be 1-5", domain.ErrInvalidArgument)
	}
	now := d.now()
	c := domain.CSATResponse{
		ID: d.newID(), TenantID: in.TenantID, CustomerID: in.CustomerID,
		TicketID: in.TicketID, ConversationID: in.ConversationID,
		Score: in.Score, Comment: in.Comment, CreatedAt: now,
	}
	if err := d.Feedback.SaveCSAT(ctx, c); err != nil {
		return domain.CSATResponse{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCSATCompleted, map[string]any{
		"score": c.Score, "customerId": c.CustomerID.String(),
	})
	return c, nil
}

// SubmitNPSInput records an NPS score.
type SubmitNPSInput struct {
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	TicketID   *uuid.UUID
	Score      int
	Comment    string
}

// SubmitNPS stores NPS feedback (score 0–10).
func (d *Deps) SubmitNPS(ctx context.Context, in SubmitNPSInput) (domain.Feedback, error) {
	if in.TenantID == uuid.Nil || in.CustomerID == uuid.Nil {
		return domain.Feedback{}, fmt.Errorf("%w: tenant_id, customer_id required", domain.ErrInvalidArgument)
	}
	if in.Score < 0 || in.Score > 10 {
		return domain.Feedback{}, fmt.Errorf("%w: nps score must be 0-10", domain.ErrInvalidArgument)
	}
	now := d.now()
	f := domain.Feedback{
		ID: d.newID(), TenantID: in.TenantID, CustomerID: in.CustomerID,
		TicketID: in.TicketID, Kind: domain.FeedbackNPS,
		Score: in.Score, Comment: in.Comment, CreatedAt: now,
	}
	if err := d.Feedback.SaveFeedback(ctx, f); err != nil {
		return domain.Feedback{}, err
	}
	d.emit(ctx, f.TenantID, f.ID, domain.EventFeedbackReceived, map[string]any{
		"kind": domain.FeedbackNPS, "score": f.Score,
	})
	return f, nil
}

// Customer360 is an aggregated customer view (read ports only).
type Customer360 struct {
	Profile  ports.ProfileSummary `json:"profile"`
	Orders   []ports.OrderSummary `json:"orders"`
	Tickets  []domain.Ticket      `json:"tickets"`
	Cases    []domain.Case        `json:"cases"`
	CSAT     []domain.CSATResponse `json:"csat"`
}

// GetCustomer360 aggregates profile + orders stubs + CRM local data.
func (d *Deps) GetCustomer360(ctx context.Context, tenantID, customerID uuid.UUID) (Customer360, error) {
	if tenantID == uuid.Nil || customerID == uuid.Nil {
		return Customer360{}, fmt.Errorf("%w: tenant_id, customer_id required", domain.ErrInvalidArgument)
	}
	out := Customer360{}
	if d.Profiles != nil {
		p, err := d.Profiles.GetProfile(ctx, tenantID, customerID)
		if err != nil {
			return Customer360{}, err
		}
		out.Profile = p
	}
	if d.Orders != nil {
		orders, err := d.Orders.ListOrders(ctx, tenantID, customerID, 10)
		if err != nil {
			return Customer360{}, err
		}
		out.Orders = orders
	}
	if d.Tickets != nil {
		out.Tickets, _ = d.Tickets.ListByCustomer(ctx, tenantID, customerID)
	}
	if d.Cases != nil {
		out.Cases, _ = d.Cases.ListByCustomer(ctx, tenantID, customerID)
	}
	if d.Feedback != nil {
		all, _ := d.Feedback.ListCSAT(ctx, tenantID)
		for _, c := range all {
			if c.CustomerID == customerID {
				out.CSAT = append(out.CSAT, c)
			}
		}
	}
	return out, nil
}

// UpsertSLAPolicyInput upserts an SLA policy.
type UpsertSLAPolicyInput struct {
	TenantID             uuid.UUID
	Name                 string
	Priority             string
	FirstResponseMinutes int
	ResolveMinutes       int
	Active               bool
}

// UpsertSLAPolicy saves an SLA policy for a priority.
func (d *Deps) UpsertSLAPolicy(ctx context.Context, in UpsertSLAPolicyInput) (domain.SLAPolicy, error) {
	if in.TenantID == uuid.Nil || !domain.ValidPriority(in.Priority) {
		return domain.SLAPolicy{}, fmt.Errorf("%w: tenant_id and valid priority required", domain.ErrInvalidArgument)
	}
	now := d.now()
	p := domain.SLAPolicy{
		ID: d.newID(), TenantID: in.TenantID, Name: in.Name, Priority: in.Priority,
		FirstResponseMinutes: in.FirstResponseMinutes, ResolveMinutes: in.ResolveMinutes,
		Active: in.Active, CreatedAt: now, UpdatedAt: now,
	}
	if existing, err := d.SLA.GetPolicyByPriority(ctx, in.TenantID, in.Priority); err == nil {
		p.ID = existing.ID
		p.CreatedAt = existing.CreatedAt
	}
	if err := d.SLA.SavePolicy(ctx, p); err != nil {
		return domain.SLAPolicy{}, err
	}
	return p, nil
}

// EvaluateSLA marks tickets past due as breached and optionally escalates.
func (d *Deps) EvaluateSLA(ctx context.Context, tenantID uuid.UUID, escalateOnBreach bool) ([]domain.Ticket, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	tickets, err := d.Tickets.ListOpen(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	now := d.now()
	var breached []domain.Ticket
	for _, t := range tickets {
		isBreach := false
		if t.ResolveDue != nil && !now.Before(*t.ResolveDue) {
			isBreach = true
		}
		if t.FirstResponseDue != nil && t.FirstRespondedAt == nil && !now.Before(*t.FirstResponseDue) {
			isBreach = true
		}
		if !isBreach || t.SLABreached {
			continue
		}
		t.SLABreached = true
		t.UpdatedAt = now
		_ = d.Tickets.Save(ctx, t)
		d.emit(ctx, t.TenantID, t.ID, domain.EventSLABreached, map[string]any{
			"priority": t.Priority, "resolveDue": t.ResolveDue,
		})
		if escalateOnBreach {
			t2, _, err := d.Escalate(ctx, EscalateTicketInput{
				TenantID: tenantID, TicketID: t.ID, Reason: "sla_breach", TriggeredBySLA: true,
			})
			if err == nil {
				t = t2
				t.SLABreached = true
				_ = d.Tickets.Save(ctx, t)
			}
		}
		breached = append(breached, t)
	}
	return breached, nil
}

// BreachEscalation is an alias that forces escalateOnBreach=true.
func (d *Deps) BreachEscalation(ctx context.Context, tenantID uuid.UUID) ([]domain.Ticket, error) {
	return d.EvaluateSLA(ctx, tenantID, true)
}

// AdminStats is aggregate CRM stats for a tenant.
type AdminStats struct {
	OpenTickets      int `json:"openTickets"`
	ActiveChats      int `json:"activeChats"`
	PublishedArticles int `json:"publishedArticles"`
	CSATCount        int `json:"csatCount"`
	AgentCount       int `json:"agentCount"`
	OpenCases        int `json:"openCases"`
}

// AdminStats returns tenant-level counters.
func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (AdminStats, error) {
	if tenantID == uuid.Nil {
		return AdminStats{}, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	stats := AdminStats{}
	if tickets, err := d.Tickets.ListOpen(ctx, tenantID); err == nil {
		stats.OpenTickets = len(tickets)
	}
	if d.Feedback != nil {
		if csat, err := d.Feedback.ListCSAT(ctx, tenantID); err == nil {
			stats.CSATCount = len(csat)
		}
	}
	if d.Agents != nil {
		if agents, err := d.Agents.ListAgents(ctx, tenantID); err == nil {
			stats.AgentCount = len(agents)
		}
	}
	if d.KB != nil {
		if arts, err := d.KB.Search(ctx, tenantID, ""); err == nil {
			stats.PublishedArticles = len(arts)
		}
	}
	return stats, nil
}
