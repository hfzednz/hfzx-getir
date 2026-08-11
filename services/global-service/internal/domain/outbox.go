package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OutboxStatusPending   = "pending"
	OutboxStatusPublished = "published"
	OutboxStatusFailed    = "failed"
	TopicGlobalEvents     = "global.events"
)

const (
	EventCountryAdded         = "CountryAdded"
	EventLanguageAdded        = "LanguageAdded"
	EventTranslationUpdated   = "TranslationUpdated"
	EventExchangeRateUpdated  = "ExchangeRateUpdated"
	EventTaxRuleUpdated       = "TaxRuleUpdated"
	EventRegionActivated      = "RegionActivated"
	EventHolidayImported      = "HolidayImported"
)

func TopicForEvent(string) string { return TopicGlobalEvents }

type OutboxMessage struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	AggregateID uuid.UUID
	Topic       string
	Key         string
	Payload     map[string]any
	Status      string
	Attempts    int
	LastError   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}
