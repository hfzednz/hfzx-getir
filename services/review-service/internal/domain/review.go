package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Target types for reviews/ratings.
const (
	TargetProduct      = "product"
	TargetOrder        = "order"
	TargetCourier      = "courier"
	TargetWarehouse    = "warehouse"
	TargetSupportAgent = "support_agent"
	TargetStore        = "store"
	TargetBrand        = "brand"
	TargetCampaign     = "campaign"
	TargetDelivery     = "delivery"
	TargetExperience   = "experience"
)

// Rating schemes.
const (
	SchemeStars5 = "stars_5"
	SchemeEmoji  = "emoji"
	SchemeThumbs = "thumbs"
)

// Review statuses.
const (
	ReviewStatusDraft             = "draft"
	ReviewStatusPendingModeration = "pending_moderation"
	ReviewStatusPublished         = "published"
	ReviewStatusRejected          = "rejected"
	ReviewStatusHidden            = "hidden"
	ReviewStatusDeleted           = "deleted"
)

// Moderation statuses.
const (
	ModerationPending  = "pending"
	ModerationApproved = "approved"
	ModerationRejected = "rejected"
	ModerationEscalated = "escalated"
)

// Quality dimensions.
const (
	DimProductQuality  = "product_quality"
	DimDeliveryQuality = "delivery_quality"
	DimPackaging       = "packaging"
	DimSupportQuality  = "support_quality"
	DimFreshness       = "freshness"
	DimAccuracy        = "accuracy"
	DimTimeliness      = "timeliness"
	DimOverall         = "overall"
)

// ValidTarget reports whether t is a supported review target type.
func ValidTarget(t string) bool {
	switch t {
	case TargetProduct, TargetOrder, TargetCourier, TargetWarehouse,
		TargetSupportAgent, TargetStore, TargetBrand, TargetCampaign,
		TargetDelivery, TargetExperience:
		return true
	default:
		return false
	}
}

// ValidScheme reports whether s is a supported rating scheme.
func ValidScheme(s string) bool {
	switch s {
	case SchemeStars5, SchemeEmoji, SchemeThumbs:
		return true
	default:
		return false
	}
}

// ValidDimension reports whether d is a quality dimension.
func ValidDimension(d string) bool {
	switch d {
	case DimProductQuality, DimDeliveryQuality, DimPackaging, DimSupportQuality,
		DimFreshness, DimAccuracy, DimTimeliness, DimOverall:
		return true
	default:
		return false
	}
}

// NormalizeStars maps emoji/thumbs into a 1–5 scale for aggregation.
func NormalizeStars(scheme string, value float64) (float64, error) {
	switch scheme {
	case SchemeStars5:
		if value < 1 || value > 5 {
			return 0, ErrInvalidArgument
		}
		return value, nil
	case SchemeEmoji:
		// 1=angry … 5=love
		if value < 1 || value > 5 {
			return 0, ErrInvalidArgument
		}
		return value, nil
	case SchemeThumbs:
		if value == 1 {
			return 5, nil
		}
		if value == 0 || value == -1 {
			return 1, nil
		}
		return 0, ErrInvalidArgument
	default:
		return 0, ErrInvalidArgument
	}
}

// Review is the aggregate root for customer feedback content.
type Review struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	AuthorID        uuid.UUID // opaque principal / customer
	TargetType      string
	TargetID        uuid.UUID
	OrderID         *uuid.UUID // optional opaque order for verified purchase
	Locale          string
	Title           string
	Body            string
	Anonymous       bool
	VerifiedPurchase bool
	VerifiedDelivery bool
	Status          string
	Sentiment       float64 // -1..1 from AI
	Topics          []string
	Tags            []string
	HelpfulCount    int
	NotHelpfulCount int
	ReportCount     int
	Pinned          bool
	Revision        int
	IdempotencyKey  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PublishedAt     *time.Time
	DeletedAt       *time.Time
}

// ValidateCreate checks create invariants.
func (r *Review) ValidateCreate() error {
	if r.TenantID == uuid.Nil || r.AuthorID == uuid.Nil || r.TargetID == uuid.Nil {
		return ErrInvalidArgument
	}
	if !ValidTarget(r.TargetType) {
		return ErrInvalidArgument
	}
	body := strings.TrimSpace(r.Body)
	if utf8.RuneCountInString(body) < 3 && strings.TrimSpace(r.Title) == "" {
		return ErrInvalidArgument
	}
	if utf8.RuneCountInString(body) > 10000 {
		return ErrInvalidArgument
	}
	return nil
}

// CanTransition reports whether a status transition is allowed.
func CanTransition(from, to string) bool {
	switch from {
	case ReviewStatusDraft:
		return to == ReviewStatusPendingModeration || to == ReviewStatusDeleted
	case ReviewStatusPendingModeration:
		return to == ReviewStatusPublished || to == ReviewStatusRejected || to == ReviewStatusHidden
	case ReviewStatusPublished:
		return to == ReviewStatusHidden || to == ReviewStatusDeleted || to == ReviewStatusPendingModeration
	case ReviewStatusRejected:
		return to == ReviewStatusPendingModeration || to == ReviewStatusDeleted
	case ReviewStatusHidden:
		return to == ReviewStatusPublished || to == ReviewStatusDeleted
	default:
		return false
	}
}

// ReviewRevision is an immutable snapshot after edit.
type ReviewRevision struct {
	ID        uuid.UUID
	ReviewID  uuid.UUID
	TenantID  uuid.UUID
	Revision  int
	Title     string
	Body      string
	Locale    string
	CreatedAt time.Time
	CreatedBy uuid.UUID
}

// Rating is a numeric/emoji/thumbs score for a target (may accompany a review).
type Rating struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	AuthorID   uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	ReviewID   *uuid.UUID
	Scheme     string
	Value      float64
	Stars      float64 // normalized 1–5
	Verified   bool
	Weight     float64
	CreatedAt  time.Time
}

// RatingAggregate is a cached score rollup for a target.
type RatingAggregate struct {
	TenantID       uuid.UUID
	TargetType     string
	TargetID       uuid.UUID
	Scheme         string
	Count          int
	SumStars       float64
	AvgStars       float64
	BayesianAvg    float64
	TimeDecayAvg   float64
	VerifiedCount  int
	VerifiedAvg    float64
	UpdatedAt      time.Time
}

// ReviewMedia is a media attachment reference (not binary SoT).
type ReviewMedia struct {
	ID           uuid.UUID
	ReviewID     uuid.UUID
	TenantID     uuid.UUID
	MediaRef     string // opaque media-service asset id / CDN key
	Kind         string // image|video|voice
	MimeType     string
	Width        int
	Height       int
	DurationMs   int
	Verified     bool
	ModerationOK bool
	CreatedAt    time.Time
}

// ReviewVote is helpful / not helpful.
type ReviewVote struct {
	ID        uuid.UUID
	ReviewID  uuid.UUID
	TenantID  uuid.UUID
	VoterID   uuid.UUID
	Helpful   bool
	CreatedAt time.Time
}

// ReviewComment is a reply thread node.
type ReviewComment struct {
	ID        uuid.UUID
	ReviewID  uuid.UUID
	TenantID  uuid.UUID
	AuthorID  uuid.UUID
	ParentID  *uuid.UUID
	Body      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReviewReport is a user/system report.
type ReviewReport struct {
	ID         uuid.UUID
	ReviewID   uuid.UUID
	TenantID   uuid.UUID
	ReporterID uuid.UUID
	Reason     string
	Details    string
	CreatedAt  time.Time
}

// QualityScore is a per-dimension score on a review.
type QualityScore struct {
	ID         uuid.UUID
	ReviewID   uuid.UUID
	TenantID   uuid.UUID
	Dimension  string
	Value      float64 // 1–5
	CreatedAt  time.Time
}

// ModerationCase tracks AI + human moderation for a review.
type ModerationCase struct {
	ID              uuid.UUID
	ReviewID        uuid.UUID
	TenantID        uuid.UUID
	Status          string
	AutoDecision    string
	AIScore         float64 // 0=safe … 1=unsafe
	Labels          []string
	FraudScore      float64
	FraudSignals    []string
	PIIMasked       bool
	AssigneeID      *uuid.UUID
	DecisionNote    string
	DecidedBy       *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DecidedAt       *time.Time
}

// TrustScore is per-reviewer trust within a tenant.
type TrustScore struct {
	TenantID           uuid.UUID
	ReviewerID         uuid.UUID
	Score              float64 // 0–100
	VerifiedPurchases  int
	PublishedReviews   int
	RejectedReviews    int
	HelpfulReceived    int
	Badges             []string
	UpdatedAt          time.Time
}

// ReputationScore is per-entity public reputation.
type ReputationScore struct {
	TenantID   uuid.UUID
	TargetType string
	TargetID   uuid.UUID
	Score      float64 // 0–100
	Tier       string  // poor|fair|good|excellent|elite
	ReviewCount int
	UpdatedAt  time.Time
}

// ReputationTier maps a 0–100 score to a display tier.
func ReputationTier(score float64) string {
	switch {
	case score >= 90:
		return "elite"
	case score >= 75:
		return "excellent"
	case score >= 60:
		return "good"
	case score >= 40:
		return "fair"
	default:
		return "poor"
	}
}

// Bayesian prior defaults.
const (
	BayesianPriorMean   = 4.0
	BayesianConfidence  = 20.0
	DefaultDecayLambda  = 0.01
)
