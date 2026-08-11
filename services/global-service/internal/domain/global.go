package domain

import (
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Money remains int64 minor units at call sites; FX converts between currencies.
type CountryStatus string

const (
	CountryDraft    CountryStatus = "draft"
	CountryActive   CountryStatus = "active"
	CountryPaused   CountryStatus = "paused"
)

type Country struct {
	ID              uuid.UUID     `json:"id"`
	TenantID        uuid.UUID     `json:"tenantId"`
	Code            string        `json:"code"` // ISO 3166-1 alpha-2
	Name            string        `json:"name"`
	DefaultLocale   string        `json:"defaultLocale"`
	DefaultCurrency string        `json:"defaultCurrency"`
	DefaultTZ       string        `json:"defaultTimezone"`
	Status          CountryStatus `json:"status"`
	DataResidency   string        `json:"dataResidency"` // region key e.g. eu-west, tr-central
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type PlaceKind string

const (
	PlaceRegion   PlaceKind = "region"
	PlaceState    PlaceKind = "state"
	PlaceCity     PlaceKind = "city"
	PlaceDistrict PlaceKind = "district"
	PlacePostal   PlaceKind = "postal"
	PlaceZone     PlaceKind = "operational_zone"
)

type Place struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	CountryID  uuid.UUID  `json:"countryId"`
	ParentID   *uuid.UUID `json:"parentId,omitempty"`
	Kind       PlaceKind  `json:"kind"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	Active     bool       `json:"active"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type Language struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Code      string    `json:"code"` // BCP47 language e.g. tr, en, ar
	Name      string    `json:"name"`
	RTL       bool      `json:"rtl"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"createdAt"`
}

type LocaleProfile struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenantId"`
	Locale         string    `json:"locale"` // e.g. tr-TR
	LanguageCode   string    `json:"languageCode"`
	CountryCode    string    `json:"countryCode"`
	DateFormat     string    `json:"dateFormat"`
	TimeFormat     string    `json:"timeFormat"`
	NumberFormat   string    `json:"numberFormat"`
	CurrencyFormat string    `json:"currencyFormat"`
	FirstDayOfWeek int       `json:"firstDayOfWeek"` // 0=Sun .. 6=Sat
	CreatedAt      time.Time `json:"createdAt"`
}

type Translation struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	Namespace  string    `json:"namespace"`
	Key        string    `json:"key"`
	Locale     string    `json:"locale"`
	Value      string    `json:"value"`
	Version    int       `json:"version"`
	Context    string    `json:"context,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type Currency struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Code        string    `json:"code"` // ISO 4217
	Name        string    `json:"name"`
	MinorUnits  int       `json:"minorUnits"`
	Symbol      string    `json:"symbol"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ExchangeRate struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	BaseCurrency string    `json:"baseCurrency"`
	QuoteCurrency string   `json:"quoteCurrency"`
	Rate         float64   `json:"rate"` // quote per 1 base
	AsOf         time.Time `json:"asOf"`
	Source       string    `json:"source"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Holiday struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	CountryID uuid.UUID `json:"countryId"`
	Date      string    `json:"date"` // YYYY-MM-DD
	Name      string    `json:"name"`
	Kind      string    `json:"kind"` // public|company
	CreatedAt time.Time `json:"createdAt"`
}

type RegionalRule struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenantId"`
	CountryID       uuid.UUID      `json:"countryId"`
	PlaceID         *uuid.UUID     `json:"placeId,omitempty"`
	MinOrderMinor   int64          `json:"minOrderMinor"`
	DeliveryFeeMinor int64         `json:"deliveryFeeMinor"`
	Currency        string         `json:"currency"`
	LegalAge        int            `json:"legalAge"`
	RestrictedSKUs  []string       `json:"restrictedSkus"`
	OpenHour        string         `json:"openHour"`  // HH:MM local
	CloseHour       string         `json:"closeHour"`
	WarehouseRules  map[string]any `json:"warehouseRules,omitempty"`
	CourierRules    map[string]any `json:"courierRules,omitempty"`
	Active          bool           `json:"active"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}

type TaxKind string

const (
	TaxVAT   TaxKind = "vat"
	TaxGST   TaxKind = "gst"
	TaxSales TaxKind = "sales_tax"
)

type TaxRule struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	CountryID  uuid.UUID  `json:"countryId"`
	PlaceID    *uuid.UUID `json:"placeId,omitempty"`
	Kind       TaxKind    `json:"kind"`
	RateBps    int        `json:"rateBps"` // 1800 = 18.00%
	Name       string     `json:"name"`
	ExemptSKUs []string   `json:"exemptSkus"`
	Active     bool       `json:"active"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type PrivacyRegime struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenantId"`
	CountryID      uuid.UUID `json:"countryId"`
	Framework      string    `json:"framework"` // GDPR|KVKK|CCPA
	ConsentRequired bool     `json:"consentRequired"`
	RetentionDays  int       `json:"retentionDays"`
	DataResidency  string    `json:"dataResidency"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// PaymentMethodAvailability configures which methods may be offered — does not charge.
type PaymentMethodAvailability struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	CountryID  uuid.UUID `json:"countryId"`
	MethodCode string    `json:"methodCode"` // card|wallet|cash|transfer|installment
	Enabled    bool      `json:"enabled"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type LogisticsPolicy struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenantId"`
	CountryID     uuid.UUID `json:"countryId"`
	SLAMinutes    int       `json:"slaMinutes"`
	HolidayRouting bool     `json:"holidayRouting"`
	ZoneCodes     []string  `json:"zoneCodes"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type LegalDocument struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	CountryID uuid.UUID `json:"countryId"`
	Kind      string    `json:"kind"` // terms|privacy|help
	Locale    string    `json:"locale"`
	Version   int       `json:"version"`
	URI       string    `json:"uri"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ResolveBundle is the client bootstrap payload for a country+locale.
type ResolveBundle struct {
	Country     Country                    `json:"country"`
	Locale      LocaleProfile              `json:"locale"`
	Currency    Currency                   `json:"currency"`
	FX          *ExchangeRate              `json:"fx,omitempty"`
	Rules       *RegionalRule              `json:"rules,omitempty"`
	Tax         []TaxRule                  `json:"tax"`
	Privacy     *PrivacyRegime             `json:"privacy,omitempty"`
	Payments    []PaymentMethodAvailability `json:"payments"`
	Logistics   *LogisticsPolicy           `json:"logistics,omitempty"`
	Translations map[string]string         `json:"translations"`
	Holidays    []Holiday                  `json:"holidays"`
}

func ValidateCountry(c Country) error {
	if c.TenantID == uuid.Nil || len(c.Code) != 2 || c.Name == "" || c.DefaultLocale == "" || c.DefaultCurrency == "" {
		return ErrInvalidArgument
	}
	return nil
}

func ValidateTranslation(t Translation) error {
	if t.TenantID == uuid.Nil || t.Namespace == "" || t.Key == "" || t.Locale == "" || t.Value == "" {
		return ErrInvalidArgument
	}
	return nil
}

// ConvertMinor converts amount from base to quote using rate (quote per 1 base).
func ConvertMinor(amountMinor int64, rate float64, fromMinor, toMinor int) int64 {
	if rate <= 0 || fromMinor < 0 || toMinor < 0 {
		return 0
	}
	major := float64(amountMinor) / math.Pow10(fromMinor)
	out := major * rate * math.Pow10(toMinor)
	return int64(math.Round(out))
}

func ResolveTranslation(primary, fallback map[string]string, key string) (string, bool) {
	if v, ok := primary[key]; ok && v != "" {
		return v, true
	}
	if v, ok := fallback[key]; ok && v != "" {
		return v, true
	}
	return "", false
}

func NormalizeLocale(locale string) string {
	return strings.ReplaceAll(strings.TrimSpace(locale), "_", "-")
}

func TaxAmountMinor(netMinor int64, rateBps int) int64 {
	if rateBps <= 0 || netMinor <= 0 {
		return 0
	}
	return (netMinor * int64(rateBps)) / 10000
}
