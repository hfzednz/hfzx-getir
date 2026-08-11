package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Article statuses.
const (
	ArticleStatusDraft     = "draft"
	ArticleStatusPublished = "published"
	ArticleStatusArchived  = "archived"
)

// Article is a knowledge-base article.
type Article struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Slug        string
	Title       string
	Body        string
	Locale      string
	Tags        []string
	Status      string
	Version     int
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Publish transitions draft → published.
func (a *Article) Publish(now time.Time) error {
	if a.Status != ArticleStatusDraft && a.Status != ArticleStatusArchived {
		return fmt.Errorf("%w: article status %s cannot publish", ErrIllegalTransition, a.Status)
	}
	a.Status = ArticleStatusPublished
	a.PublishedAt = &now
	a.UpdatedAt = now
	return nil
}

// ArticleVersion stores a historical KB body snapshot.
type ArticleVersion struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	ArticleID uuid.UUID
	Version   int
	Title     string
	Body      string
	CreatedAt time.Time
}
