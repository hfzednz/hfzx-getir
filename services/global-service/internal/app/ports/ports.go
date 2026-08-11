package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type CountryRepo interface {
	Save(ctx context.Context, c domain.Country) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Country, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Country, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Country, error)
}

type PlaceRepo interface {
	Save(ctx context.Context, p domain.Place) error
	ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.Place, error)
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Place, error)
}

type LanguageRepo interface {
	Save(ctx context.Context, l domain.Language) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Language, error)
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Language, error)
}

type LocaleRepo interface {
	Save(ctx context.Context, l domain.LocaleProfile) error
	Get(ctx context.Context, tenantID uuid.UUID, locale string) (domain.LocaleProfile, error)
}

type TranslationRepo interface {
	Save(ctx context.Context, t domain.Translation) error
	ListNamespace(ctx context.Context, tenantID uuid.UUID, ns, locale string) ([]domain.Translation, error)
	Get(ctx context.Context, tenantID uuid.UUID, ns, key, locale string) (domain.Translation, error)
}

type CurrencyRepo interface {
	Save(ctx context.Context, c domain.Currency) error
	GetByCode(ctx context.Context, tenantID uuid.UUID, code string) (domain.Currency, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Currency, error)
}

type RateRepo interface {
	Save(ctx context.Context, r domain.ExchangeRate) error
	Latest(ctx context.Context, tenantID uuid.UUID, base, quote string) (domain.ExchangeRate, error)
	History(ctx context.Context, tenantID uuid.UUID, base, quote string, limit int) ([]domain.ExchangeRate, error)
}

type HolidayRepo interface {
	Save(ctx context.Context, h domain.Holiday) error
	ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.Holiday, error)
}

type RuleRepo interface {
	Save(ctx context.Context, r domain.RegionalRule) error
	GetForCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.RegionalRule, error)
}

type TaxRepo interface {
	Save(ctx context.Context, t domain.TaxRule) error
	ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.TaxRule, error)
}

type PrivacyRepo interface {
	Save(ctx context.Context, p domain.PrivacyRegime) error
	GetByCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.PrivacyRegime, error)
}

type PayAvailRepo interface {
	Save(ctx context.Context, p domain.PaymentMethodAvailability) error
	ListByCountry(ctx context.Context, tenantID, countryID uuid.UUID) ([]domain.PaymentMethodAvailability, error)
}

type LogisticsRepo interface {
	Save(ctx context.Context, p domain.LogisticsPolicy) error
	GetByCountry(ctx context.Context, tenantID, countryID uuid.UUID) (domain.LogisticsPolicy, error)
}

type LegalRepo interface {
	Save(ctx context.Context, d domain.LegalDocument) error
	List(ctx context.Context, tenantID, countryID uuid.UUID, locale string) ([]domain.LegalDocument, error)
}

type FXClient interface {
	FetchRate(ctx context.Context, base, quote string) (float64, error)
}

type AITranslateClient interface {
	Translate(ctx context.Context, text, from, to string) (string, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}
