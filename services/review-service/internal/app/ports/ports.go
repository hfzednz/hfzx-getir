package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/domain"
)

// Clock abstracts time.
type Clock interface {
	Now() time.Time
}

// IDGen abstracts UUID generation.
type IDGen interface {
	New() uuid.UUID
}

// ReviewRepo persists reviews and revisions.
type ReviewRepo interface {
	Save(ctx context.Context, r domain.Review) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Review, error)
	GetByIdempotency(ctx context.Context, tenantID uuid.UUID, key string) (domain.Review, bool, error)
	ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, status string, limit int) ([]domain.Review, error)
	ListByAuthor(ctx context.Context, tenantID, authorID uuid.UUID, limit int) ([]domain.Review, error)
	SaveRevision(ctx context.Context, rev domain.ReviewRevision) error
	ListRevisions(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewRevision, error)
	CountRecentByAuthor(ctx context.Context, tenantID, authorID uuid.UUID, since time.Time) (int, error)
	CountDupBody(ctx context.Context, tenantID uuid.UUID, bodyHash string, since time.Time) (int, error)
}

// RatingRepo persists ratings and aggregates.
type RatingRepo interface {
	Save(ctx context.Context, r domain.Rating) error
	ListByTarget(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) ([]domain.Rating, error)
	GetAggregate(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID, scheme string) (domain.RatingAggregate, error)
	SaveAggregate(ctx context.Context, a domain.RatingAggregate) error
}

// MediaRepo persists media refs.
type MediaRepo interface {
	Save(ctx context.Context, m domain.ReviewMedia) error
	ListByReview(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewMedia, error)
}

// InteractionRepo votes, comments, reports.
type InteractionRepo interface {
	SaveVote(ctx context.Context, v domain.ReviewVote) error
	GetVote(ctx context.Context, reviewID, voterID uuid.UUID) (domain.ReviewVote, bool, error)
	SaveComment(ctx context.Context, c domain.ReviewComment) error
	ListComments(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.ReviewComment, error)
	SaveReport(ctx context.Context, r domain.ReviewReport) error
}

// QualityRepo persists quality dimensions.
type QualityRepo interface {
	Save(ctx context.Context, q domain.QualityScore) error
	ListByReview(ctx context.Context, tenantID, reviewID uuid.UUID) ([]domain.QualityScore, error)
}

// ModerationRepo persists moderation cases.
type ModerationRepo interface {
	Save(ctx context.Context, m domain.ModerationCase) error
	GetByReview(ctx context.Context, tenantID, reviewID uuid.UUID) (domain.ModerationCase, error)
	ListPending(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.ModerationCase, error)
}

// TrustRepo persists reviewer trust.
type TrustRepo interface {
	Get(ctx context.Context, tenantID, reviewerID uuid.UUID) (domain.TrustScore, error)
	Save(ctx context.Context, t domain.TrustScore) error
}

// ReputationRepo persists entity reputation.
type ReputationRepo interface {
	Get(ctx context.Context, tenantID uuid.UUID, targetType string, targetID uuid.UUID) (domain.ReputationScore, error)
	Save(ctx context.Context, r domain.ReputationScore) error
}

// OutboxRepository transactional outbox.
type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

// EventPublisher publishes domain events.
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// SearchIndexer indexes published reviews.
type SearchIndexer interface {
	IndexReview(ctx context.Context, r domain.Review) error
	DeleteReview(ctx context.Context, tenantID, reviewID uuid.UUID) error
	Search(ctx context.Context, tenantID uuid.UUID, query string, targetType string, limit int) ([]uuid.UUID, error)
}

// OrderReadClient verifies purchase/delivery (opaque).
type OrderReadClient interface {
	VerifyPurchase(ctx context.Context, tenantID, customerID, orderID, targetID uuid.UUID, targetType string) (purchased bool, delivered bool, err error)
}

// MediaClient registers media refs / moderation status.
type MediaClient interface {
	ValidateRef(ctx context.Context, tenantID uuid.UUID, mediaRef string) (ok bool, kind string, err error)
}

// ModerationClient AI content analysis.
type ModerationClient interface {
	Analyze(ctx context.Context, tenantID uuid.UUID, title, body string, locale string) (ModerationResult, error)
	Summarize(ctx context.Context, tenantID uuid.UUID, bodies []string) (string, error)
	ExtractTopics(ctx context.Context, tenantID uuid.UUID, body string) ([]string, float64, error)
}

// ModerationResult is AI moderation output.
type ModerationResult struct {
	UnsafeScore float64
	Labels      []string
	PIIFound    bool
	MaskedBody  string
	Sentiment   float64
}
