package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TemplateStatus is the lifecycle of a template.
type TemplateStatus string

const (
	TemplateDraft    TemplateStatus = "draft"
	TemplateApproved TemplateStatus = "approved"
	TemplateActive   TemplateStatus = "active"
	TemplateRetired  TemplateStatus = "retired"
)

// Template is a multi-version notification template.
type Template struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Channel   Channel
	Locale    string
	Version   int
	Status    TemplateStatus
	Subject   string
	Body      string
	Variant   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

var varPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.]+)\s*\}\}`)

// Render substitutes {{var}} placeholders from vars. Missing keys become empty.
func Render(body string, vars map[string]string) string {
	if vars == nil {
		vars = map[string]string{}
	}
	return varPattern.ReplaceAllStringFunc(body, func(m string) string {
		sub := varPattern.FindStringSubmatch(m)
		if len(sub) < 2 {
			return m
		}
		key := strings.TrimSpace(sub[1])
		if v, ok := vars[key]; ok {
			return v
		}
		return ""
	})
}

// Preview renders subject + body.
func (t Template) Preview(vars map[string]string) (subject, body string) {
	return Render(t.Subject, vars), Render(t.Body, vars)
}
