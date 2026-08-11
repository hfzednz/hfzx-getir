package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/domain"
)

// CreateTicketInput creates a support ticket.
type CreateTicketInput struct {
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	Subject        string
	Description    string
	Priority       string
	Category       string
	IdempotencyKey string
	Tags           []string
	ActorID        *uuid.UUID
}

// CreateTicket creates a ticket or returns the existing one for the same idempotency key.
func (d *Deps) CreateTicket(ctx context.Context, in CreateTicketInput) (domain.Ticket, error) {
	if in.TenantID == uuid.Nil || in.CustomerID == uuid.Nil || strings.TrimSpace(in.Subject) == "" {
		return domain.Ticket{}, fmt.Errorf("%w: tenant_id, customer_id, subject required", domain.ErrInvalidArgument)
	}
	priority := in.Priority
	if priority == "" {
		priority = domain.PriorityNormal
	}
	if !domain.ValidPriority(priority) {
		return domain.Ticket{}, fmt.Errorf("%w: invalid priority", domain.ErrInvalidArgument)
	}
	category := in.Category
	if category == "" {
		category = domain.CategoryOther
	}
	if !domain.ValidCategory(category) {
		return domain.Ticket{}, fmt.Errorf("%w: invalid category", domain.ErrInvalidArgument)
	}
	if in.IdempotencyKey != "" {
		existing, err := d.Tickets.GetByIdempotencyKey(ctx, in.TenantID, in.IdempotencyKey)
		if err == nil {
			return existing, nil
		}
		if err != domain.ErrNotFound {
			return domain.Ticket{}, err
		}
	}

	now := d.now()
	t := domain.Ticket{
		ID: d.newID(), TenantID: in.TenantID, CustomerID: in.CustomerID,
		Status: domain.TicketStatusOpen, Priority: priority, Category: category,
		Subject: strings.TrimSpace(in.Subject), Description: in.Description,
		IdempotencyKey: in.IdempotencyKey, Tags: in.Tags,
		CreatedAt: now, UpdatedAt: now,
	}

	if d.SLA != nil {
		if pol, err := d.SLA.GetPolicyByPriority(ctx, in.TenantID, priority); err == nil && pol.Active {
			fr, rd := pol.ComputeDueTimes(now)
			t.FirstResponseDue = &fr
			t.ResolveDue = &rd
		} else {
			// default SLA: 60m first response, 24h resolve
			fr := now.Add(60 * time.Minute)
			rd := now.Add(24 * time.Hour)
			t.FirstResponseDue = &fr
			t.ResolveDue = &rd
		}
	}

	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventCreated, Payload: map[string]any{"subject": t.Subject},
		CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketCreated, map[string]any{
		"customerId": t.CustomerID.String(), "priority": t.Priority, "category": t.Category,
	})
	if d.Notify != nil {
		_ = d.Notify.Notify(ctx, t.TenantID, t.CustomerID, "support.ticket_created", map[string]any{
			"ticketId": t.ID.String(), "subject": t.Subject,
		})
	}
	return t, nil
}

// AssignTicketInput assigns an agent to a ticket.
type AssignTicketInput struct {
	TenantID uuid.UUID
	TicketID uuid.UUID
	AgentID  uuid.UUID
	TeamID   *uuid.UUID
	ActorID  *uuid.UUID
}

// AssignTicket assigns an agent and moves ticket toward in_progress.
func (d *Deps) AssignTicket(ctx context.Context, in AssignTicketInput) (domain.Ticket, error) {
	if in.TenantID == uuid.Nil || in.TicketID == uuid.Nil || in.AgentID == uuid.Nil {
		return domain.Ticket{}, fmt.Errorf("%w: tenant_id, ticket_id, agent_id required", domain.ErrInvalidArgument)
	}
	t, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := d.now()
	t.AssigneeID = &in.AgentID
	t.TeamID = in.TeamID
	if t.Status == domain.TicketStatusOpen || t.Status == domain.TicketStatusPending || t.Status == domain.TicketStatusReopened {
		if err := t.Transition(domain.TicketStatusInProgress, now); err != nil {
			return domain.Ticket{}, err
		}
	} else {
		t.UpdatedAt = now
	}
	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventAssigned, Payload: map[string]any{"agentId": in.AgentID.String()},
		CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketAssigned, map[string]any{"agentId": in.AgentID.String()})
	return t, nil
}

// AddNoteInput adds an internal note.
type AddNoteInput struct {
	TenantID uuid.UUID
	TicketID uuid.UUID
	AuthorID uuid.UUID
	Body     string
}

// AddNote appends an internal ticket note.
func (d *Deps) AddNote(ctx context.Context, in AddNoteInput) (domain.TicketNote, error) {
	if in.TenantID == uuid.Nil || in.TicketID == uuid.Nil || strings.TrimSpace(in.Body) == "" {
		return domain.TicketNote{}, fmt.Errorf("%w: tenant_id, ticket_id, body required", domain.ErrInvalidArgument)
	}
	if _, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID); err != nil {
		return domain.TicketNote{}, err
	}
	now := d.now()
	n := domain.TicketNote{
		ID: d.newID(), TenantID: in.TenantID, TicketID: in.TicketID,
		AuthorID: in.AuthorID, Body: strings.TrimSpace(in.Body), CreatedAt: now,
	}
	if err := d.Tickets.AddNote(ctx, n); err != nil {
		return domain.TicketNote{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: in.TicketID, ActorID: &in.AuthorID,
		Type: domain.TicketEventNote, Payload: map[string]any{"noteId": n.ID.String()},
		CreatedAt: now,
	})
	return n, nil
}

// EscalateTicketInput escalates a ticket.
type EscalateTicketInput struct {
	TenantID       uuid.UUID
	TicketID       uuid.UUID
	Reason         string
	ActorID        *uuid.UUID
	TriggeredBySLA bool
}

// Escalate raises priority and records an escalation.
func (d *Deps) Escalate(ctx context.Context, in EscalateTicketInput) (domain.Ticket, domain.Escalation, error) {
	if in.TenantID == uuid.Nil || in.TicketID == uuid.Nil {
		return domain.Ticket{}, domain.Escalation{}, fmt.Errorf("%w: tenant_id, ticket_id required", domain.ErrInvalidArgument)
	}
	t, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID)
	if err != nil {
		return domain.Ticket{}, domain.Escalation{}, err
	}
	now := d.now()
	from := t.Priority
	to := domain.NextPriority(from)
	t.Priority = to
	t.UpdatedAt = now
	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, domain.Escalation{}, err
	}
	esc := domain.Escalation{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID,
		FromPriority: from, ToPriority: to, Reason: in.Reason,
		TriggeredBySLA: in.TriggeredBySLA, CreatedAt: now,
	}
	if d.SLA != nil {
		_ = d.SLA.SaveEscalation(ctx, esc)
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventEscalated,
		Payload: map[string]any{"from": from, "to": to, "reason": in.Reason},
		CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketEscalated, map[string]any{
		"fromPriority": from, "toPriority": to, "reason": in.Reason,
	})
	return t, esc, nil
}

// ResolveTicketInput resolves a ticket.
type ResolveTicketInput struct {
	TenantID uuid.UUID
	TicketID uuid.UUID
	ActorID  *uuid.UUID
	Note     string
}

// Resolve transitions ticket to resolved.
func (d *Deps) Resolve(ctx context.Context, in ResolveTicketInput) (domain.Ticket, error) {
	t, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := d.now()
	if err := t.Transition(domain.TicketStatusResolved, now); err != nil {
		return domain.Ticket{}, err
	}
	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventResolved, Payload: map[string]any{"note": in.Note},
		CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketResolved, nil)
	return t, nil
}

// CloseTicketInput closes a resolved ticket.
type CloseTicketInput struct {
	TenantID uuid.UUID
	TicketID uuid.UUID
	ActorID  *uuid.UUID
}

// Close closes a ticket. Only allowed from resolved (illegal from open).
func (d *Deps) Close(ctx context.Context, in CloseTicketInput) (domain.Ticket, error) {
	t, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := d.now()
	if err := t.Transition(domain.TicketStatusClosed, now); err != nil {
		return domain.Ticket{}, err
	}
	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventClosed, CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketClosed, nil)
	return t, nil
}

// ReopenTicketInput reopens a closed/resolved ticket.
type ReopenTicketInput struct {
	TenantID uuid.UUID
	TicketID uuid.UUID
	ActorID  *uuid.UUID
	Reason   string
}

// Reopen reopens a closed or resolved ticket.
func (d *Deps) Reopen(ctx context.Context, in ReopenTicketInput) (domain.Ticket, error) {
	t, err := d.Tickets.Get(ctx, in.TenantID, in.TicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := d.now()
	target := domain.TicketStatusReopened
	if err := t.Transition(target, now); err != nil {
		return domain.Ticket{}, err
	}
	if err := d.Tickets.Save(ctx, t); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: t.ID, ActorID: in.ActorID,
		Type: domain.TicketEventReopened, Payload: map[string]any{"reason": in.Reason},
		CreatedAt: now,
	})
	d.emit(ctx, t.TenantID, t.ID, domain.EventTicketReopened, map[string]any{"reason": in.Reason})
	return t, nil
}

// MergeTicketsInput merges source into target.
type MergeTicketsInput struct {
	TenantID       uuid.UUID
	SourceTicketID uuid.UUID
	TargetTicketID uuid.UUID
	ActorID        *uuid.UUID
}

// MergeTickets closes source into target.
func (d *Deps) MergeTickets(ctx context.Context, in MergeTicketsInput) (domain.Ticket, error) {
	if in.SourceTicketID == in.TargetTicketID {
		return domain.Ticket{}, fmt.Errorf("%w: cannot merge ticket into itself", domain.ErrInvalidArgument)
	}
	src, err := d.Tickets.Get(ctx, in.TenantID, in.SourceTicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	tgt, err := d.Tickets.Get(ctx, in.TenantID, in.TargetTicketID)
	if err != nil {
		return domain.Ticket{}, err
	}
	now := d.now()
	src.MergedIntoID = &tgt.ID
	if src.Status != domain.TicketStatusClosed {
		if src.Status != domain.TicketStatusResolved {
			_ = src.Transition(domain.TicketStatusResolved, now)
		}
		_ = src.Transition(domain.TicketStatusClosed, now)
	} else {
		src.UpdatedAt = now
	}
	if err := d.Tickets.Save(ctx, src); err != nil {
		return domain.Ticket{}, err
	}
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: src.ID, ActorID: in.ActorID,
		Type: domain.TicketEventMerged, Payload: map[string]any{"into": tgt.ID.String()},
		CreatedAt: now,
	})
	_ = d.Tickets.AddEvent(ctx, domain.TicketEvent{
		ID: d.newID(), TenantID: in.TenantID, TicketID: tgt.ID, ActorID: in.ActorID,
		Type: domain.TicketEventMerged, Payload: map[string]any{"from": src.ID.String()},
		CreatedAt: now,
	})
	return tgt, nil
}
