package memory

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Countries    map[uuid.UUID]domain.Country
	CountryByCode map[string]uuid.UUID
	Places       map[uuid.UUID]domain.Place
	Languages    map[uuid.UUID]domain.Language
	Locales      map[string]domain.LocaleProfile
	Translations map[string]domain.Translation
	Currencies   map[string]domain.Currency
	Rates        []domain.ExchangeRate
	Holidays     map[uuid.UUID]domain.Holiday
	Rules        map[uuid.UUID]domain.RegionalRule // by country
	Taxes        map[uuid.UUID]domain.TaxRule
	Privacy      map[uuid.UUID]domain.PrivacyRegime // by country
	PayAvail     map[uuid.UUID]domain.PaymentMethodAvailability
	Logistics    map[uuid.UUID]domain.LogisticsPolicy
	Legal        map[uuid.UUID]domain.LegalDocument
	Outbox       map[uuid.UUID]domain.OutboxMessage
	Cache        map[string]string
}

func NewStore() *Store {
	return &Store{
		Countries: map[uuid.UUID]domain.Country{}, CountryByCode: map[string]uuid.UUID{},
		Places: map[uuid.UUID]domain.Place{}, Languages: map[uuid.UUID]domain.Language{},
		Locales: map[string]domain.LocaleProfile{}, Translations: map[string]domain.Translation{},
		Currencies: map[string]domain.Currency{}, Rates: []domain.ExchangeRate{},
		Holidays: map[uuid.UUID]domain.Holiday{}, Rules: map[uuid.UUID]domain.RegionalRule{},
		Taxes: map[uuid.UUID]domain.TaxRule{}, Privacy: map[uuid.UUID]domain.PrivacyRegime{},
		PayAvail: map[uuid.UUID]domain.PaymentMethodAvailability{}, Logistics: map[uuid.UUID]domain.LogisticsPolicy{},
		Legal: map[uuid.UUID]domain.LegalDocument{}, Outbox: map[uuid.UUID]domain.OutboxMessage{},
		Cache: map[string]string{},
	}
}

func ck(tenantID uuid.UUID, code string) string { return tenantID.String() + ":" + code }
func tk(tenantID uuid.UUID, ns, key, locale string) string {
	return tenantID.String() + ":" + ns + ":" + key + ":" + locale
}
func lk(tenantID uuid.UUID, locale string) string { return tenantID.String() + ":" + locale }

type Repos struct {
	Countries    *CountryRepo
	Places       *PlaceRepo
	Languages    *LanguageRepo
	Locales      *LocaleRepo
	Translations *TranslationRepo
	Currencies   *CurrencyRepo
	Rates        *RateRepo
	Holidays     *HolidayRepo
	Rules        *RuleRepo
	Taxes        *TaxRepo
	Privacy      *PrivacyRepo
	PayAvail     *PayAvailRepo
	Logistics    *LogisticsRepo
	Legal        *LegalRepo
	Outbox       *OutboxRepo
	FX           *MockFX
	AI           *MockAI
	Metrics      *MockMetrics
	Cache        *CacheRepo
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Countries: &CountryRepo{s: s}, Places: &PlaceRepo{s: s}, Languages: &LanguageRepo{s: s},
		Locales: &LocaleRepo{s: s}, Translations: &TranslationRepo{s: s}, Currencies: &CurrencyRepo{s: s},
		Rates: &RateRepo{s: s}, Holidays: &HolidayRepo{s: s}, Rules: &RuleRepo{s: s}, Taxes: &TaxRepo{s: s},
		Privacy: &PrivacyRepo{s: s}, PayAvail: &PayAvailRepo{s: s}, Logistics: &LogisticsRepo{s: s},
		Legal: &LegalRepo{s: s}, Outbox: &OutboxRepo{s: s}, FX: &MockFX{}, AI: &MockAI{},
		Metrics: &MockMetrics{}, Cache: &CacheRepo{s: s},
	}
}

type CountryRepo struct{ s *Store }

func (r *CountryRepo) Save(_ context.Context, c domain.Country) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Countries[c.ID] = c
	r.s.CountryByCode[ck(c.TenantID, c.Code)] = c.ID
	return nil
}
func (r *CountryRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Country, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Countries[id]
	if !ok || c.TenantID != tenantID {
		return domain.Country{}, domain.ErrNotFound
	}
	return c, nil
}
func (r *CountryRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Country, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.CountryByCode[ck(tenantID, code)]
	if !ok {
		return domain.Country{}, domain.ErrNotFound
	}
	return r.s.Countries[id], nil
}
func (r *CountryRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Country, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Country{}
	for _, c := range r.s.Countries {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type PlaceRepo struct{ s *Store }

func (r *PlaceRepo) Save(_ context.Context, p domain.Place) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Places[p.ID] = p
	return nil
}
func (r *PlaceRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Place, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Places[id]
	if !ok || p.TenantID != tenantID {
		return domain.Place{}, domain.ErrNotFound
	}
	return p, nil
}
func (r *PlaceRepo) ListByCountry(_ context.Context, tenantID, countryID uuid.UUID) ([]domain.Place, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Place{}
	for _, p := range r.s.Places {
		if p.TenantID == tenantID && p.CountryID == countryID {
			out = append(out, p)
		}
	}
	return out, nil
}

type LanguageRepo struct{ s *Store }

func (r *LanguageRepo) Save(_ context.Context, l domain.Language) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Languages[l.ID] = l
	return nil
}
func (r *LanguageRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Language, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, l := range r.s.Languages {
		if l.TenantID == tenantID && l.Code == code {
			return l, nil
		}
	}
	return domain.Language{}, domain.ErrNotFound
}
func (r *LanguageRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Language, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Language{}
	for _, l := range r.s.Languages {
		if l.TenantID == tenantID {
			out = append(out, l)
		}
	}
	return out, nil
}

type LocaleRepo struct{ s *Store }

func (r *LocaleRepo) Save(_ context.Context, l domain.LocaleProfile) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Locales[lk(l.TenantID, l.Locale)] = l
	return nil
}
func (r *LocaleRepo) Get(_ context.Context, tenantID uuid.UUID, locale string) (domain.LocaleProfile, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	l, ok := r.s.Locales[lk(tenantID, locale)]
	if !ok {
		return domain.LocaleProfile{}, domain.ErrNotFound
	}
	return l, nil
}

type TranslationRepo struct{ s *Store }

func (r *TranslationRepo) Save(_ context.Context, t domain.Translation) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Translations[tk(t.TenantID, t.Namespace, t.Key, t.Locale)] = t
	return nil
}
func (r *TranslationRepo) Get(_ context.Context, tenantID uuid.UUID, ns, key, locale string) (domain.Translation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	t, ok := r.s.Translations[tk(tenantID, ns, key, locale)]
	if !ok {
		return domain.Translation{}, domain.ErrNotFound
	}
	return t, nil
}
func (r *TranslationRepo) ListNamespace(_ context.Context, tenantID uuid.UUID, ns, locale string) ([]domain.Translation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Translation{}
	for _, t := range r.s.Translations {
		if t.TenantID == tenantID && t.Namespace == ns && t.Locale == locale {
			out = append(out, t)
		}
	}
	return out, nil
}

type CurrencyRepo struct{ s *Store }

func (r *CurrencyRepo) Save(_ context.Context, c domain.Currency) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Currencies[ck(c.TenantID, c.Code)] = c
	return nil
}
func (r *CurrencyRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Currency, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Currencies[ck(tenantID, code)]
	if !ok {
		return domain.Currency{}, domain.ErrNotFound
	}
	return c, nil
}
func (r *CurrencyRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Currency, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Currency{}
	for _, c := range r.s.Currencies {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type RateRepo struct{ s *Store }

func (r *RateRepo) Save(_ context.Context, rate domain.ExchangeRate) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Rates = append(r.s.Rates, rate)
	return nil
}
func (r *RateRepo) Latest(_ context.Context, tenantID uuid.UUID, base, quote string) (domain.ExchangeRate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	var best domain.ExchangeRate
	found := false
	for _, x := range r.s.Rates {
		if x.TenantID == tenantID && x.BaseCurrency == base && x.QuoteCurrency == quote {
			if !found || x.AsOf.After(best.AsOf) {
				best = x
				found = true
			}
		}
	}
	if !found {
		return domain.ExchangeRate{}, domain.ErrNotFound
	}
	return best, nil
}
func (r *RateRepo) History(_ context.Context, tenantID uuid.UUID, base, quote string, limit int) ([]domain.ExchangeRate, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ExchangeRate{}
	for i := len(r.s.Rates) - 1; i >= 0; i-- {
		x := r.s.Rates[i]
		if x.TenantID == tenantID && x.BaseCurrency == base && x.QuoteCurrency == quote {
			out = append(out, x)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type HolidayRepo struct{ s *Store }

func (r *HolidayRepo) Save(_ context.Context, h domain.Holiday) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Holidays[h.ID] = h
	return nil
}
func (r *HolidayRepo) ListByCountry(_ context.Context, tenantID, countryID uuid.UUID) ([]domain.Holiday, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Holiday{}
	for _, h := range r.s.Holidays {
		if h.TenantID == tenantID && h.CountryID == countryID {
			out = append(out, h)
		}
	}
	return out, nil
}

type RuleRepo struct{ s *Store }

func (r *RuleRepo) Save(_ context.Context, rule domain.RegionalRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Rules[rule.CountryID] = rule
	return nil
}
func (r *RuleRepo) GetForCountry(_ context.Context, tenantID, countryID uuid.UUID) (domain.RegionalRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	rule, ok := r.s.Rules[countryID]
	if !ok || rule.TenantID != tenantID {
		return domain.RegionalRule{}, domain.ErrNotFound
	}
	return rule, nil
}

type TaxRepo struct{ s *Store }

func (r *TaxRepo) Save(_ context.Context, t domain.TaxRule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Taxes[t.ID] = t
	return nil
}
func (r *TaxRepo) ListByCountry(_ context.Context, tenantID, countryID uuid.UUID) ([]domain.TaxRule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.TaxRule{}
	for _, t := range r.s.Taxes {
		if t.TenantID == tenantID && t.CountryID == countryID {
			out = append(out, t)
		}
	}
	return out, nil
}

type PrivacyRepo struct{ s *Store }

func (r *PrivacyRepo) Save(_ context.Context, p domain.PrivacyRegime) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Privacy[p.CountryID] = p
	return nil
}
func (r *PrivacyRepo) GetByCountry(_ context.Context, tenantID, countryID uuid.UUID) (domain.PrivacyRegime, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Privacy[countryID]
	if !ok || p.TenantID != tenantID {
		return domain.PrivacyRegime{}, domain.ErrNotFound
	}
	return p, nil
}

type PayAvailRepo struct{ s *Store }

func (r *PayAvailRepo) Save(_ context.Context, p domain.PaymentMethodAvailability) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.PayAvail[p.ID] = p
	return nil
}
func (r *PayAvailRepo) ListByCountry(_ context.Context, tenantID, countryID uuid.UUID) ([]domain.PaymentMethodAvailability, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.PaymentMethodAvailability{}
	for _, p := range r.s.PayAvail {
		if p.TenantID == tenantID && p.CountryID == countryID {
			out = append(out, p)
		}
	}
	return out, nil
}

type LogisticsRepo struct{ s *Store }

func (r *LogisticsRepo) Save(_ context.Context, p domain.LogisticsPolicy) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Logistics[p.CountryID] = p
	return nil
}
func (r *LogisticsRepo) GetByCountry(_ context.Context, tenantID, countryID uuid.UUID) (domain.LogisticsPolicy, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	p, ok := r.s.Logistics[countryID]
	if !ok || p.TenantID != tenantID {
		return domain.LogisticsPolicy{}, domain.ErrNotFound
	}
	return p, nil
}

type LegalRepo struct{ s *Store }

func (r *LegalRepo) Save(_ context.Context, d domain.LegalDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Legal[d.ID] = d
	return nil
}
func (r *LegalRepo) List(_ context.Context, tenantID, countryID uuid.UUID, locale string) ([]domain.LegalDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.LegalDocument{}
	for _, d := range r.s.Legal {
		if d.TenantID == tenantID && d.CountryID == countryID && (locale == "" || d.Locale == locale) {
			out = append(out, d)
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}
func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}

type CacheRepo struct{ s *Store }

func (r *CacheRepo) Get(_ context.Context, key string) (string, bool, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	v, ok := r.s.Cache[key]
	return v, ok, nil
}
func (r *CacheRepo) Set(_ context.Context, key, value string, _ time.Duration) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Cache[key] = value
	return nil
}

type MockFX struct{}

func (MockFX) FetchRate(_ context.Context, _, _ string) (float64, error) { return 34.2, nil }

type MockAI struct{}

func (MockAI) Translate(_ context.Context, text, _, to string) (string, error) {
	return "[" + to + "] " + text, nil
}

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
