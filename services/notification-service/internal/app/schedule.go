package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// ScheduleSendInput creates a delayed send.
type ScheduleSendInput struct {
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	Channel        domain.Channel
	Priority       domain.Priority
	TemplateKey    string
	Locale         string
	Recipient      string
	Subject        string
	Body           string
	Vars           map[string]string
	IdempotencyKey string
	SendAt         time.Time
}

// ScheduleSend enqueues a future send.
func (d *Deps) ScheduleSend(ctx context.Context, in ScheduleSendInput) (domain.Schedule, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil || !in.Channel.Valid() || in.SendAt.IsZero() {
		return domain.Schedule{}, fmt.Errorf("%w: tenant_id, principal_id, channel, send_at required", domain.ErrInvalidArgument)
	}
	priority := in.Priority
	if priority == "" {
		priority = domain.PriorityTransactional
	}
	locale := in.Locale
	if locale == "" {
		locale = "en"
	}
	now := d.now()
	s := domain.Schedule{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		Channel: in.Channel, Priority: priority, TemplateKey: in.TemplateKey,
		Locale: locale, Recipient: in.Recipient, Subject: in.Subject, Body: in.Body,
		Vars: in.Vars, IdempotencyKey: in.IdempotencyKey, SendAt: in.SendAt.UTC(),
		Status: domain.SchedulePending, CreatedAt: now, UpdatedAt: now,
	}
	if err := d.Schedules.Create(ctx, s); err != nil {
		return domain.Schedule{}, err
	}
	return s, nil
}

// CancelSchedule cancels a pending schedule.
func (d *Deps) CancelSchedule(ctx context.Context, tenantID, scheduleID uuid.UUID) (domain.Schedule, error) {
	if tenantID == uuid.Nil || scheduleID == uuid.Nil {
		return domain.Schedule{}, fmt.Errorf("%w: tenant_id and schedule_id required", domain.ErrInvalidArgument)
	}
	return d.Schedules.Cancel(ctx, tenantID, scheduleID, d.now())
}

// ProcessDueSchedules dispatches due pending schedules.
func (d *Deps) ProcessDueSchedules(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 100
	}
	due, err := d.Schedules.ListDue(ctx, d.now(), limit)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, s := range due {
		msg, err := d.Send(ctx, SendInput{
			TenantID: s.TenantID, PrincipalID: s.PrincipalID,
			Channel: s.Channel, Priority: s.Priority, TemplateKey: s.TemplateKey,
			Locale: s.Locale, Recipient: s.Recipient, Subject: s.Subject, Body: s.Body,
			Vars: s.Vars, IdempotencyKey: s.IdempotencyKey,
		})
		now := d.now()
		s.UpdatedAt = now
		if err != nil && msg.ID == uuid.Nil {
			// hard failure before message created — leave pending for retry? mark processed with no message
			continue
		}
		mid := msg.ID
		s.MessageID = &mid
		s.Status = domain.ScheduleProcessed
		_ = d.Schedules.Update(ctx, s)
		n++
	}
	return n, nil
}
