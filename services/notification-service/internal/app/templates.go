package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/notification-service/internal/domain"
)

// UpsertTemplateInput creates or updates a template version.
type UpsertTemplateInput struct {
	TenantID uuid.UUID
	ID       uuid.UUID // optional; generated if nil
	Key      string
	Channel  domain.Channel
	Locale   string
	Subject  string
	Body     string
	Variant  string
	Activate bool
}

// UpsertTemplate upserts a template (draft or active).
func (d *Deps) UpsertTemplate(ctx context.Context, in UpsertTemplateInput) (domain.Template, error) {
	if in.TenantID == uuid.Nil || strings.TrimSpace(in.Key) == "" || !in.Channel.Valid() {
		return domain.Template{}, fmt.Errorf("%w: tenant_id, key, channel required", domain.ErrInvalidArgument)
	}
	locale := in.Locale
	if locale == "" {
		locale = "en"
	}
	now := d.now()
	version := 1
	if existing, err := d.Templates.GetByKey(ctx, in.TenantID, in.Key, in.Channel, locale); err == nil {
		version = existing.Version + 1
	}
	id := in.ID
	if id == uuid.Nil {
		id = d.newID()
	}
	status := domain.TemplateDraft
	if in.Activate {
		status = domain.TemplateActive
	}
	t := domain.Template{
		ID: id, TenantID: in.TenantID, Key: in.Key, Channel: in.Channel,
		Locale: locale, Version: version, Status: status,
		Subject: in.Subject, Body: in.Body, Variant: in.Variant,
		CreatedAt: now, UpdatedAt: now,
	}
	return d.Templates.Upsert(ctx, t)
}

// ApproveTemplateInput approves a template.
type ApproveTemplateInput struct {
	TenantID   uuid.UUID
	TemplateID uuid.UUID
}

// ApproveTemplate marks a template approved/active.
func (d *Deps) ApproveTemplate(ctx context.Context, in ApproveTemplateInput) (domain.Template, error) {
	if in.TenantID == uuid.Nil || in.TemplateID == uuid.Nil {
		return domain.Template{}, fmt.Errorf("%w: tenant_id and template_id required", domain.ErrInvalidArgument)
	}
	return d.Templates.Approve(ctx, in.TenantID, in.TemplateID, d.now())
}

// PreviewTemplateInput previews a template render.
type PreviewTemplateInput struct {
	TenantID   uuid.UUID
	TemplateID uuid.UUID
	Key        string
	Channel    domain.Channel
	Locale     string
	Vars       map[string]string
}

// PreviewResult holds rendered subject/body.
type PreviewResult struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// PreviewTemplate renders {{var}} substitution without sending.
func (d *Deps) PreviewTemplate(ctx context.Context, in PreviewTemplateInput) (PreviewResult, error) {
	var t domain.Template
	var err error
	if in.TemplateID != uuid.Nil {
		t, err = d.Templates.Get(ctx, in.TenantID, in.TemplateID)
	} else if in.Key != "" && in.Channel.Valid() {
		locale := in.Locale
		if locale == "" {
			locale = "en"
		}
		t, err = d.Templates.GetByKey(ctx, in.TenantID, in.Key, in.Channel, locale)
	} else {
		return PreviewResult{}, fmt.Errorf("%w: template_id or key+channel required", domain.ErrInvalidArgument)
	}
	if err != nil {
		return PreviewResult{}, err
	}
	subj, body := t.Preview(in.Vars)
	return PreviewResult{Subject: subj, Body: body}, nil
}
