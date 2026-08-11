package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/domain"
)

// Store is the in-memory aggregate store for dev mode.
type Store struct {
	mu          sync.RWMutex
	Reviews     map[uuid.UUID]domain.Review
	Idem        map[string]uuid.UUID
	Revisions   []domain.ReviewRevision
	Ratings     []domain.Rating
	Aggregates  map[string]domain.RatingAggregate
	Media       []domain.ReviewMedia
	Votes       map[string]domain.ReviewVote
	Comments    []domain.ReviewComment
	Reports     []domain.ReviewReport
	Quality     []domain.QualityScore
	Moderation  map[uuid.UUID]domain.ModerationCase // by reviewID (latest)
	Trust       map[string]domain.TrustScore
	Reputation  map[string]domain.ReputationScore
	Outbox      []domain.OutboxMessage
	BodyHashes  map[string][]time.Time // hash -> created times
	SearchIndex map[uuid.UUID]domain.Review
}

// NewStore creates an empty store.
func NewStore() *Store {
	return &Store{
		Reviews:     make(map[uuid.UUID]domain.Review),
		Idem:        make(map[string]uuid.UUID),
		Aggregates:  make(map[string]domain.RatingAggregate),
		Votes:       make(map[string]domain.ReviewVote),
		Moderation:  make(map[uuid.UUID]domain.ModerationCase),
		Trust:       make(map[string]domain.TrustScore),
		Reputation:  make(map[string]domain.ReputationScore),
		BodyHashes:  make(map[string][]time.Time),
		SearchIndex: make(map[uuid.UUID]domain.Review),
	}
}

func aggKey(tenantID uuid.UUID, targetType string, targetID uuid.UUID, scheme string) string {
	return tenantID.String() + "|" + targetType + "|" + targetID.String() + "|" + scheme
}

func trustKey(tenantID, reviewerID uuid.UUID) string {
	return tenantID.String() + "|" + reviewerID.String()
}

func repKey(tenantID uuid.UUID, targetType string, targetID uuid.UUID) string {
	return tenantID.String() + "|" + targetType + "|" + targetID.String()
}

func voteKey(reviewID, voterID uuid.UUID) string {
	return reviewID.String() + "|" + voterID.String()
}
