package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/crm-service/internal/domain"
)

// AIAssistInput asks the AI assistant for a reply.
type AIAssistInput struct {
	TenantID       uuid.UUID
	ConversationID uuid.UUID
	CustomerID     uuid.UUID
	Text           string
	AutoEscalate   bool
}

// AIAssistResult is the AI assist outcome.
type AIAssistResult struct {
	Intent     string             `json:"intent"`
	Reply      string             `json:"reply"`
	Confidence float64            `json:"confidence"`
	Sentiment  string             `json:"sentiment"`
	Sources    []string           `json:"sources"`
	Escalated  bool               `json:"escalated"`
	Ticket     *domain.Ticket     `json:"ticket,omitempty"`
	Message    *domain.Message    `json:"message,omitempty"`
	KBHits     []domain.Article   `json:"kbHits,omitempty"`
}

// AIAssist retrieves KB context, drafts an LLM reply, and escalates on low confidence or negative sentiment.
func (d *Deps) AIAssist(ctx context.Context, in AIAssistInput) (AIAssistResult, error) {
	if in.TenantID == uuid.Nil || strings.TrimSpace(in.Text) == "" {
		return AIAssistResult{}, fmt.Errorf("%w: tenant_id, text required", domain.ErrInvalidArgument)
	}
	if d.LLM == nil {
		return AIAssistResult{}, fmt.Errorf("%w: llm client not configured", domain.ErrInvariant)
	}

	intent, err := d.LLM.DetectIntent(ctx, in.Text)
	if err != nil {
		return AIAssistResult{}, err
	}
	sentiment, err := d.LLM.AnalyzeSentiment(ctx, in.Text)
	if err != nil {
		return AIAssistResult{}, err
	}

	var snippets []string
	var hits []domain.Article
	if d.KB != nil {
		hits, _ = d.KB.Search(ctx, in.TenantID, in.Text)
		for _, a := range hits {
			snippets = append(snippets, a.Title+": "+a.Body)
			if len(snippets) >= 3 {
				break
			}
		}
	}

	reply, err := d.LLM.DraftReply(ctx, in.Text, snippets)
	if err != nil {
		return AIAssistResult{}, err
	}

	result := AIAssistResult{
		Intent: intent.Intent, Reply: reply.Reply, Confidence: reply.Confidence,
		Sentiment: sentiment, Sources: reply.Sources, KBHits: hits,
	}

	escalate := reply.Confidence < d.confidenceThreshold() || sentiment == domain.SentimentNegative
	if escalate && in.AutoEscalate {
		result.Escalated = true
		customerID := in.CustomerID
		if customerID == uuid.Nil && in.ConversationID != uuid.Nil {
			if c, err := d.Chats.GetConversation(ctx, in.TenantID, in.ConversationID); err == nil {
				customerID = c.CustomerID
			}
		}
		if customerID != uuid.Nil {
			t, err := d.CreateTicket(ctx, CreateTicketInput{
				TenantID: in.TenantID, CustomerID: customerID,
				Subject: "AI escalation: " + intent.Intent,
				Description: in.Text, Priority: domain.PriorityHigh, Category: domain.CategoryComplaint,
			})
			if err == nil {
				t2, _, escErr := d.Escalate(ctx, EscalateTicketInput{
					TenantID: in.TenantID, TicketID: t.ID, Reason: "ai_low_confidence_or_negative_sentiment",
				})
				if escErr == nil {
					result.Ticket = &t2
				} else {
					result.Ticket = &t
				}
			}
		}
	} else if escalate {
		result.Escalated = true
	}

	if in.ConversationID != uuid.Nil && d.Chats != nil {
		msg, _, err := d.PostMessage(ctx, PostMessageInput{
			TenantID: in.TenantID, ConversationID: in.ConversationID,
			SenderRole: domain.SenderAI, Body: reply.Reply,
		})
		if err == nil {
			result.Message = &msg
		}
	}

	return result, nil
}

// UpsertArticleInput creates or updates a KB article.
type UpsertArticleInput struct {
	TenantID  uuid.UUID
	ArticleID uuid.UUID // optional for update
	Slug      string
	Title     string
	Body      string
	Locale    string
	Tags      []string
}

// UpsertArticle saves a draft article (or bumps version on update).
func (d *Deps) UpsertArticle(ctx context.Context, in UpsertArticleInput) (domain.Article, error) {
	if in.TenantID == uuid.Nil || strings.TrimSpace(in.Slug) == "" || strings.TrimSpace(in.Title) == "" {
		return domain.Article{}, fmt.Errorf("%w: tenant_id, slug, title required", domain.ErrInvalidArgument)
	}
	now := d.now()
	locale := in.Locale
	if locale == "" {
		locale = "en"
	}
	var a domain.Article
	if in.ArticleID != uuid.Nil {
		existing, err := d.KB.GetArticle(ctx, in.TenantID, in.ArticleID)
		if err != nil {
			return domain.Article{}, err
		}
		a = existing
		a.Title = strings.TrimSpace(in.Title)
		a.Body = in.Body
		a.Tags = in.Tags
		a.Locale = locale
		a.Version++
		a.UpdatedAt = now
		a.Status = domain.ArticleStatusDraft
		_ = d.KB.SaveVersion(ctx, domain.ArticleVersion{
			ID: d.newID(), TenantID: in.TenantID, ArticleID: a.ID,
			Version: a.Version, Title: a.Title, Body: a.Body, CreatedAt: now,
		})
	} else {
		if existing, err := d.KB.GetBySlug(ctx, in.TenantID, in.Slug); err == nil {
			a = existing
			a.Title = strings.TrimSpace(in.Title)
			a.Body = in.Body
			a.Tags = in.Tags
			a.Locale = locale
			a.Version++
			a.UpdatedAt = now
			a.Status = domain.ArticleStatusDraft
		} else {
			a = domain.Article{
				ID: d.newID(), TenantID: in.TenantID, Slug: strings.TrimSpace(in.Slug),
				Title: strings.TrimSpace(in.Title), Body: in.Body, Locale: locale,
				Tags: in.Tags, Status: domain.ArticleStatusDraft, Version: 1,
				CreatedAt: now, UpdatedAt: now,
			}
		}
	}
	if err := d.KB.SaveArticle(ctx, a); err != nil {
		return domain.Article{}, err
	}
	return a, nil
}

// PublishArticle publishes a draft article.
func (d *Deps) PublishArticle(ctx context.Context, tenantID, articleID uuid.UUID) (domain.Article, error) {
	a, err := d.KB.GetArticle(ctx, tenantID, articleID)
	if err != nil {
		return domain.Article{}, err
	}
	now := d.now()
	if err := a.Publish(now); err != nil {
		return domain.Article{}, err
	}
	if err := d.KB.SaveArticle(ctx, a); err != nil {
		return domain.Article{}, err
	}
	d.emit(ctx, a.TenantID, a.ID, domain.EventArticlePublished, map[string]any{"slug": a.Slug})
	return a, nil
}

// SearchKB searches published articles.
func (d *Deps) SearchKB(ctx context.Context, tenantID uuid.UUID, query string) ([]domain.Article, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("%w: tenant_id required", domain.ErrInvalidArgument)
	}
	return d.KB.Search(ctx, tenantID, query)
}
