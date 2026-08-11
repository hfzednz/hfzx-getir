package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const maxModelVersionLen = 64

// AICustomerModel holds scored customer features for personalization / CRM.
// Heavy training and inference live outside this service.
type AICustomerModel struct {
	ProfileID             uuid.UUID
	Frequency             float64
	AvgOrderValueMinor    int64
	ChurnProb             float64
	PreferredOrderHours   []int
	PreferredOrderWeekdays []int
	PriceSensitivity      float64
	BrandAffinity         map[string]any
	CategoryAffinity      map[string]any
	ModelVersion          string
	UpdatedAt             time.Time
	CreatedAt             time.Time
}

// Validate checks structural invariants and score bounds.
func (m AICustomerModel) Validate() error {
	if m.ProfileID == uuid.Nil {
		return fmt.Errorf("%w: profile_id required", ErrInvalidArgument)
	}
	if m.Frequency < 0 {
		return fmt.Errorf("%w: frequency must be >= 0", ErrInvalidArgument)
	}
	if m.AvgOrderValueMinor < 0 {
		return fmt.Errorf("%w: avg_order_value_minor must be >= 0", ErrInvalidArgument)
	}
	if m.ChurnProb < 0 || m.ChurnProb > 1 {
		return fmt.Errorf("%w: churn_prob must be in [0,1]", ErrInvalidArgument)
	}
	if m.PriceSensitivity < 0 || m.PriceSensitivity > 1 {
		return fmt.Errorf("%w: price_sensitivity must be in [0,1]", ErrInvalidArgument)
	}
	for _, h := range m.PreferredOrderHours {
		if h < 0 || h > 23 {
			return fmt.Errorf("%w: preferred_order_hours must be in [0,23]", ErrInvalidArgument)
		}
	}
	for _, d := range m.PreferredOrderWeekdays {
		if d < 0 || d > 6 {
			return fmt.Errorf("%w: preferred_order_weekdays must be in [0,6]", ErrInvalidArgument)
		}
	}
	if utf8.RuneCountInString(m.ModelVersion) > maxModelVersionLen {
		return fmt.Errorf("%w: model_version too long", ErrInvalidArgument)
	}
	return nil
}
