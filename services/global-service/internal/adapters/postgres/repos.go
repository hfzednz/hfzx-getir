package postgres

import "database/sql"

// Repos groups global-service persistence adapters.
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
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Countries: &CountryRepo{DB: db}, Places: &PlaceRepo{DB: db}, Languages: &LanguageRepo{DB: db},
		Locales: &LocaleRepo{DB: db}, Translations: &TranslationRepo{DB: db}, Currencies: &CurrencyRepo{DB: db},
		Rates: &RateRepo{DB: db}, Holidays: &HolidayRepo{DB: db}, Rules: &RuleRepo{DB: db}, Taxes: &TaxRepo{DB: db},
		Privacy: &PrivacyRepo{DB: db}, PayAvail: &PayAvailRepo{DB: db}, Logistics: &LogisticsRepo{DB: db},
		Legal: &LegalRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
