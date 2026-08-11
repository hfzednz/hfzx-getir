package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/app/ports"
	"github.com/nexora/notification-service/internal/domain"
)

// SendInput is a single notification send request.
type SendInput struct {
	TenantID       uuid.UUID
	PrincipalID    uuid.UUID
	OrderID        *uuid.UUID
	Channel        domain.Channel
	Priority       domain.Priority
	TemplateKey    string
	Locale         string
	Recipient      string
	Subject        string
	Body           string
	Vars           map[string]string
	IdempotencyKey string
}

// Send respects preferences, renders template, enqueues and dispatches.
func (d *Deps) Send(ctx context.Context, in SendInput) (domain.Message, error) {
	if in.TenantID == uuid.Nil || in.PrincipalID == uuid.Nil || !in.Channel.Valid() {
		return domain.Message{}, fmt.Errorf("%w: tenant_id, principal_id, channel required", domain.ErrInvalidArgument)
	}
	priority := in.Priority
	if priority == "" {
		priority = domain.PriorityTransactional
	}
	if !priority.Valid() {
		return domain.Message{}, fmt.Errorf("%w: invalid priority", domain.ErrInvalidArgument)
	}
	if in.IdempotencyKey != "" {
		if existing, err := d.Messages.GetByIdempotency(ctx, in.TenantID, in.IdempotencyKey); err == nil {
			return existing, nil
		}
	}

	locale := in.Locale
	if locale == "" {
		locale = "en"
	}
	now := d.now()
	msg := domain.Message{
		ID: d.newID(), TenantID: in.TenantID, PrincipalID: in.PrincipalID,
		OrderID: in.OrderID, Channel: in.Channel, Priority: priority,
		TemplateKey: in.TemplateKey, Locale: locale, Recipient: in.Recipient,
		Subject: in.Subject, Body: in.Body, Vars: in.Vars,
		IdempotencyKey: in.IdempotencyKey, Status: domain.MessageQueued,
		MaxAttempts: domain.DefaultMaxAttempts, CreatedAt: now, UpdatedAt: now,
	}

	// Preferences
	pref, _ := d.GetPreferences(ctx, in.TenantID, in.PrincipalID)
	if !pref.Allow(in.Channel, priority, now) {
		msg.Status = domain.MessageSuppressed
		optedOut := pref.ChannelOptOut != nil && pref.ChannelOptOut[in.Channel] && !priority.OverridesMarketingOptOut()
		if optedOut {
			msg.SuppressReason = "opt_out"
		} else {
			msg.SuppressReason = "quiet_hours"
		}
		if err := d.Messages.Create(ctx, msg); err != nil {
			return domain.Message{}, err
		}
		d.emit(ctx, msg, domain.EventNotificationQueued, map[string]any{"suppressed": true, "reason": msg.SuppressReason})
		return msg, nil
	}

	// Render template if key provided
	if in.TemplateKey != "" {
		t, err := d.Templates.GetByKey(ctx, in.TenantID, in.TemplateKey, in.Channel, locale)
		if err != nil {
			return domain.Message{}, fmt.Errorf("%w: template %s", err, in.TemplateKey)
		}
		subj, body := t.Preview(in.Vars)
		msg.Subject = subj
		msg.Body = body
		tid := t.ID
		msg.TemplateID = &tid
	} else {
		if msg.Subject == "" && in.Subject != "" {
			msg.Subject = domain.Render(in.Subject, in.Vars)
		} else if in.Vars != nil && msg.Subject != "" {
			msg.Subject = domain.Render(msg.Subject, in.Vars)
		}
		if msg.Body == "" && in.Body != "" {
			msg.Body = domain.Render(in.Body, in.Vars)
		} else if in.Vars != nil && msg.Body != "" {
			msg.Body = domain.Render(msg.Body, in.Vars)
		}
	}

	if err := d.Messages.Create(ctx, msg); err != nil {
		return domain.Message{}, err
	}
	d.emit(ctx, msg, domain.EventNotificationQueued, nil)
	return d.dispatchMessage(ctx, msg)
}

// SendBulkInput sends to many principals.
type SendBulkInput struct {
	TenantID       uuid.UUID
	PrincipalIDs   []uuid.UUID
	Channel        domain.Channel
	Priority       domain.Priority
	TemplateKey    string
	Locale         string
	RecipientFn    func(principalID uuid.UUID) string
	Vars           map[string]string
	IdempotencyKey string // base key; per-principal suffix applied
}

// SendBulk sends one notification per principal.
func (d *Deps) SendBulk(ctx context.Context, in SendBulkInput) ([]domain.Message, error) {
	if in.TenantID == uuid.Nil || len(in.PrincipalIDs) == 0 || !in.Channel.Valid() {
		return nil, fmt.Errorf("%w: tenant_id, principal_ids, channel required", domain.ErrInvalidArgument)
	}
	out := make([]domain.Message, 0, len(in.PrincipalIDs))
	for _, pid := range in.PrincipalIDs {
		recipient := ""
		if in.RecipientFn != nil {
			recipient = in.RecipientFn(pid)
		}
		idem := in.IdempotencyKey
		if idem != "" {
			idem = idem + ":" + pid.String()
		}
		msg, err := d.Send(ctx, SendInput{
			TenantID: in.TenantID, PrincipalID: pid, Channel: in.Channel,
			Priority: in.Priority, TemplateKey: in.TemplateKey, Locale: in.Locale,
			Recipient: recipient, Vars: in.Vars, IdempotencyKey: idem,
		})
		if err != nil {
			return out, err
		}
		out = append(out, msg)
	}
	return out, nil
}

func (d *Deps) dispatchMessage(ctx context.Context, msg domain.Message) (domain.Message, error) {
	now := d.now()
	msg.Status = domain.MessageSending
	msg.UpdatedAt = now
	msg.Attempts++
	_ = d.Messages.Update(ctx, msg)

	providerName, ref, err := d.sendToProvider(ctx, msg)
	attempt := domain.DeliveryAttempt{
		ID: d.newID(), TenantID: msg.TenantID, MessageID: msg.ID,
		AttemptNo: msg.Attempts, Provider: providerName, CreatedAt: now,
	}
	if err != nil {
		attempt.Status = "failed"
		attempt.Error = err.Error()
		_ = d.Deliveries.CreateAttempt(ctx, attempt)
		msg.LastError = err.Error()
		msg.Provider = providerName
		msg.UpdatedAt = d.now()
		if msg.Attempts >= msg.MaxAttempts {
			msg.Status = domain.MessageFailed
			_ = d.Messages.Update(ctx, msg)
			d.emit(ctx, msg, domain.EventNotificationFailed, map[string]any{"error": err.Error()})
			_, _ = d.MoveToDLQ(ctx, MoveToDLQInput{TenantID: msg.TenantID, MessageID: msg.ID, Reason: "max_retries"})
			return msg, fmt.Errorf("%w: %v", domain.ErrProviderFailed, err)
		}
		msg.Status = domain.MessageFailed
		_ = d.Messages.Update(ctx, msg)
		d.emit(ctx, msg, domain.EventNotificationFailed, map[string]any{"error": err.Error(), "retriable": true})
		return msg, fmt.Errorf("%w: %v", domain.ErrProviderFailed, err)
	}

	attempt.Status = "success"
	attempt.ProviderRef = ref
	_ = d.Deliveries.CreateAttempt(ctx, attempt)

	sentAt := d.now()
	msg.Status = domain.MessageSent
	msg.Provider = providerName
	msg.ProviderRef = ref
	msg.SentAt = &sentAt
	msg.UpdatedAt = sentAt
	msg.LastError = ""
	_ = d.Messages.Update(ctx, msg)

	if msg.Channel == domain.ChannelInApp {
		_ = d.Inbox.Create(ctx, domain.InboxItem{
			ID: d.newID(), TenantID: msg.TenantID, PrincipalID: msg.PrincipalID,
			MessageID: msg.ID, Title: msg.Subject, Body: msg.Body,
			CreatedAt: sentAt,
		})
	}

	d.emit(ctx, msg, domain.EventNotificationSent, map[string]any{"provider": providerName, "providerRef": ref})
	return msg, nil
}

func (d *Deps) sendToProvider(ctx context.Context, msg domain.Message) (provider, ref string, err error) {
	switch msg.Channel {
	case domain.ChannelEmail:
		if d.Email == nil {
			return "smtp", "", fmt.Errorf("email provider not configured")
		}
		to := msg.Recipient
		res, e := d.Email.Send(ctx, ports.EmailSendRequest{
			TenantID: msg.TenantID, To: to, Subject: msg.Subject, Body: msg.Body,
		})
		return "smtp", res.ProviderRef, e
	case domain.ChannelSMS:
		if d.SMS == nil {
			return "sms", "", fmt.Errorf("sms provider not configured")
		}
		res, e := d.SMS.Send(ctx, ports.SMSSendRequest{
			TenantID: msg.TenantID, To: msg.Recipient, Body: msg.Body,
		})
		return "sms", res.ProviderRef, e
	case domain.ChannelWhatsApp:
		if d.WhatsApp == nil {
			return "whatsapp", "", fmt.Errorf("whatsapp provider not configured")
		}
		res, e := d.WhatsApp.Send(ctx, ports.WhatsAppSendRequest{
			TenantID: msg.TenantID, To: msg.Recipient, Body: msg.Body,
		})
		return "whatsapp", res.ProviderRef, e
	case domain.ChannelPush, domain.ChannelWeb:
		if d.Push == nil {
			return "push", "", fmt.Errorf("push provider not configured")
		}
		token := msg.Recipient
		platform := domain.PlatformFCM
		if token == "" {
			devs, e := d.Devices.ListActive(ctx, msg.TenantID, msg.PrincipalID)
			if e != nil || len(devs) == 0 {
				return "push", "", fmt.Errorf("%w: no device token", domain.ErrNotFound)
			}
			token = devs[0].Token
			platform = devs[0].Platform
		}
		res, e := d.Push.Send(ctx, ports.PushSendRequest{
			TenantID: msg.TenantID, PrincipalID: msg.PrincipalID,
			Token: token, Platform: platform, Title: msg.Subject, Body: msg.Body, Data: msg.Vars,
		})
		return string(platform), res.ProviderRef, e
	case domain.ChannelInApp:
		return "in_app", "inbox:" + msg.ID.String(), nil
	default:
		return "", "", fmt.Errorf("%w: unsupported channel %s", domain.ErrInvalidArgument, msg.Channel)
	}
}
