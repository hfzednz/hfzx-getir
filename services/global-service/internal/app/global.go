package app

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/domain"
)

func (d *Deps) UpsertCountry(ctx context.Context, c domain.Country) (domain.Country, error) {
	if err := domain.ValidateCountry(c); err != nil {
		return domain.Country{}, err
	}
	c.Code = strings.ToUpper(c.Code)
	now := d.now()
	if c.ID == uuid.Nil {
		if existing, err := d.Countries.GetByCode(ctx, c.TenantID, c.Code); err == nil {
			c.ID = existing.ID
			c.CreatedAt = existing.CreatedAt
		} else {
			c.ID = d.newID()
			c.CreatedAt = now
		}
	}
	if c.Status == "" {
		c.Status = domain.CountryDraft
	}
	c.UpdatedAt = now
	if err := d.Countries.Save(ctx, c); err != nil {
		return domain.Country{}, err
	}
	d.emit(ctx, c.TenantID, c.ID, domain.EventCountryAdded, map[string]any{"code": c.Code})
	return c, nil
}

func (d *Deps) ActivateCountry(ctx context.Context, tenantID, id uuid.UUID) (domain.Country, error) {
	c, err := d.Countries.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Country{}, err
	}
	c.Status = domain.CountryActive
	c.UpdatedAt = d.now()
	if err := d.Countries.Save(ctx, c); err != nil {
		return domain.Country{}, err
	}
	d.emit(ctx, tenantID, c.ID, domain.EventRegionActivated, map[string]any{"code": c.Code})
	return c, nil
}

func (d *Deps) UpsertPlace(ctx context.Context, p domain.Place) (domain.Place, error) {
	if p.TenantID == uuid.Nil || p.CountryID == uuid.Nil || p.Kind == "" || p.Code == "" || p.Name == "" {
		return domain.Place{}, domain.ErrInvalidArgument
	}
	if _, err := d.Countries.Get(ctx, p.TenantID, p.CountryID); err != nil {
		return domain.Place{}, err
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
		p.CreatedAt = d.now()
	}
	p.Active = true
	if err := d.Places.Save(ctx, p); err != nil {
		return domain.Place{}, err
	}
	return p, nil
}

func (d *Deps) AddLanguage(ctx context.Context, l domain.Language) (domain.Language, error) {
	if l.TenantID == uuid.Nil || l.Code == "" || l.Name == "" {
		return domain.Language{}, domain.ErrInvalidArgument
	}
	l.ID = d.newID()
	l.Enabled = true
	l.CreatedAt = d.now()
	if err := d.Languages.Save(ctx, l); err != nil {
		return domain.Language{}, err
	}
	d.emit(ctx, l.TenantID, l.ID, domain.EventLanguageAdded, map[string]any{"code": l.Code, "rtl": l.RTL})
	return l, nil
}

func (d *Deps) UpsertLocale(ctx context.Context, loc domain.LocaleProfile) (domain.LocaleProfile, error) {
	if loc.TenantID == uuid.Nil || loc.Locale == "" {
		return domain.LocaleProfile{}, domain.ErrInvalidArgument
	}
	loc.Locale = domain.NormalizeLocale(loc.Locale)
	if loc.ID == uuid.Nil {
		if existing, err := d.Locales.Get(ctx, loc.TenantID, loc.Locale); err == nil {
			loc.ID = existing.ID
			loc.CreatedAt = existing.CreatedAt
		} else {
			loc.ID = d.newID()
			loc.CreatedAt = d.now()
		}
	}
	if loc.DateFormat == "" {
		loc.DateFormat = "YYYY-MM-DD"
	}
	if loc.TimeFormat == "" {
		loc.TimeFormat = "HH:mm"
	}
	if err := d.Locales.Save(ctx, loc); err != nil {
		return domain.LocaleProfile{}, err
	}
	return loc, nil
}

func (d *Deps) UpsertTranslation(ctx context.Context, t domain.Translation) (domain.Translation, error) {
	if err := domain.ValidateTranslation(t); err != nil {
		return domain.Translation{}, err
	}
	t.Locale = domain.NormalizeLocale(t.Locale)
	if existing, err := d.Translations.Get(ctx, t.TenantID, t.Namespace, t.Key, t.Locale); err == nil {
		t.ID = existing.ID
		t.Version = existing.Version + 1
	} else {
		t.ID = d.newID()
		t.Version = 1
	}
	t.UpdatedAt = d.now()
	if err := d.Translations.Save(ctx, t); err != nil {
		return domain.Translation{}, err
	}
	d.emit(ctx, t.TenantID, t.ID, domain.EventTranslationUpdated, map[string]any{
		"namespace": t.Namespace, "key": t.Key, "locale": t.Locale, "version": t.Version,
	})
	return t, nil
}

func (d *Deps) AIAssistTranslate(ctx context.Context, tenantID uuid.UUID, ns, key, fromLocale, toLocale, source string) (domain.Translation, error) {
	text := source
	if d.AI != nil {
		if out, err := d.AI.Translate(ctx, source, fromLocale, toLocale); err == nil && out != "" {
			text = out
		}
	}
	return d.UpsertTranslation(ctx, domain.Translation{
		TenantID: tenantID, Namespace: ns, Key: key, Locale: toLocale, Value: text, Context: "ai_assisted",
	})
}

func (d *Deps) UpsertCurrency(ctx context.Context, c domain.Currency) (domain.Currency, error) {
	if c.TenantID == uuid.Nil || len(c.Code) != 3 {
		return domain.Currency{}, domain.ErrInvalidArgument
	}
	c.Code = strings.ToUpper(c.Code)
	if c.ID == uuid.Nil {
		if existing, err := d.Currencies.GetByCode(ctx, c.TenantID, c.Code); err == nil {
			c.ID = existing.ID
			c.CreatedAt = existing.CreatedAt
		} else {
			c.ID = d.newID()
			c.CreatedAt = d.now()
		}
	}
	if c.MinorUnits <= 0 {
		c.MinorUnits = 2
	}
	c.Enabled = true
	if err := d.Currencies.Save(ctx, c); err != nil {
		return domain.Currency{}, err
	}
	return c, nil
}

func (d *Deps) UpsertExchangeRate(ctx context.Context, r domain.ExchangeRate) (domain.ExchangeRate, error) {
	if r.TenantID == uuid.Nil || r.BaseCurrency == "" || r.QuoteCurrency == "" || r.Rate <= 0 {
		return domain.ExchangeRate{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	r.BaseCurrency = strings.ToUpper(r.BaseCurrency)
	r.QuoteCurrency = strings.ToUpper(r.QuoteCurrency)
	if r.AsOf.IsZero() {
		r.AsOf = d.now()
	}
	if r.Source == "" {
		r.Source = "manual"
	}
	r.CreatedAt = d.now()
	if err := d.Rates.Save(ctx, r); err != nil {
		return domain.ExchangeRate{}, err
	}
	d.emit(ctx, r.TenantID, r.ID, domain.EventExchangeRateUpdated, map[string]any{
		"base": r.BaseCurrency, "quote": r.QuoteCurrency, "rate": r.Rate,
	})
	return r, nil
}

func (d *Deps) RefreshFX(ctx context.Context, tenantID uuid.UUID, base, quote string) (domain.ExchangeRate, error) {
	rate := 1.0
	src := "stub"
	if d.FX != nil {
		if v, err := d.FX.FetchRate(ctx, base, quote); err == nil && v > 0 {
			rate = v
			src = "feed"
		}
	}
	return d.UpsertExchangeRate(ctx, domain.ExchangeRate{
		TenantID: tenantID, BaseCurrency: base, QuoteCurrency: quote, Rate: rate, Source: src,
	})
}

func (d *Deps) Convert(ctx context.Context, tenantID uuid.UUID, amountMinor int64, from, to string) (int64, error) {
	if from == to {
		return amountMinor, nil
	}
	fromC, err := d.Currencies.GetByCode(ctx, tenantID, from)
	if err != nil {
		return 0, err
	}
	toC, err := d.Currencies.GetByCode(ctx, tenantID, to)
	if err != nil {
		return 0, err
	}
	r, err := d.Rates.Latest(ctx, tenantID, strings.ToUpper(from), strings.ToUpper(to))
	if err != nil {
		return 0, err
	}
	return domain.ConvertMinor(amountMinor, r.Rate, fromC.MinorUnits, toC.MinorUnits), nil
}

func (d *Deps) ImportHoliday(ctx context.Context, h domain.Holiday) (domain.Holiday, error) {
	if h.TenantID == uuid.Nil || h.CountryID == uuid.Nil || h.Date == "" || h.Name == "" {
		return domain.Holiday{}, domain.ErrInvalidArgument
	}
	h.ID = d.newID()
	h.CreatedAt = d.now()
	if h.Kind == "" {
		h.Kind = "public"
	}
	if err := d.Holidays.Save(ctx, h); err != nil {
		return domain.Holiday{}, err
	}
	d.emit(ctx, h.TenantID, h.ID, domain.EventHolidayImported, map[string]any{"date": h.Date, "name": h.Name})
	return h, nil
}

func (d *Deps) UpsertRule(ctx context.Context, r domain.RegionalRule) (domain.RegionalRule, error) {
	if r.TenantID == uuid.Nil || r.CountryID == uuid.Nil || r.Currency == "" {
		return domain.RegionalRule{}, domain.ErrInvalidArgument
	}
	if r.ID == uuid.Nil {
		if existing, err := d.Rules.GetForCountry(ctx, r.TenantID, r.CountryID); err == nil {
			r.ID = existing.ID
		} else {
			r.ID = d.newID()
		}
	}
	r.Active = true
	r.UpdatedAt = d.now()
	if err := d.Rules.Save(ctx, r); err != nil {
		return domain.RegionalRule{}, err
	}
	return r, nil
}

func (d *Deps) UpsertTax(ctx context.Context, t domain.TaxRule) (domain.TaxRule, error) {
	if t.TenantID == uuid.Nil || t.CountryID == uuid.Nil || t.Kind == "" || t.RateBps < 0 {
		return domain.TaxRule{}, domain.ErrInvalidArgument
	}
	if t.ID == uuid.Nil {
		t.ID = d.newID()
	}
	t.Active = true
	t.UpdatedAt = d.now()
	if err := d.Taxes.Save(ctx, t); err != nil {
		return domain.TaxRule{}, err
	}
	d.emit(ctx, t.TenantID, t.ID, domain.EventTaxRuleUpdated, map[string]any{"kind": string(t.Kind), "rateBps": t.RateBps})
	return t, nil
}

func (d *Deps) UpsertPrivacy(ctx context.Context, p domain.PrivacyRegime) (domain.PrivacyRegime, error) {
	if p.TenantID == uuid.Nil || p.CountryID == uuid.Nil || p.Framework == "" {
		return domain.PrivacyRegime{}, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		if existing, err := d.Privacy.GetByCountry(ctx, p.TenantID, p.CountryID); err == nil {
			p.ID = existing.ID
		} else {
			p.ID = d.newID()
		}
	}
	p.UpdatedAt = d.now()
	if err := d.Privacy.Save(ctx, p); err != nil {
		return domain.PrivacyRegime{}, err
	}
	return p, nil
}

func (d *Deps) UpsertPayAvail(ctx context.Context, p domain.PaymentMethodAvailability) (domain.PaymentMethodAvailability, error) {
	if p.TenantID == uuid.Nil || p.CountryID == uuid.Nil || p.MethodCode == "" {
		return domain.PaymentMethodAvailability{}, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	p.UpdatedAt = d.now()
	if err := d.PayAvail.Save(ctx, p); err != nil {
		return domain.PaymentMethodAvailability{}, err
	}
	return p, nil
}

func (d *Deps) UpsertLogistics(ctx context.Context, p domain.LogisticsPolicy) (domain.LogisticsPolicy, error) {
	if p.TenantID == uuid.Nil || p.CountryID == uuid.Nil {
		return domain.LogisticsPolicy{}, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		if existing, err := d.Logistics.GetByCountry(ctx, p.TenantID, p.CountryID); err == nil {
			p.ID = existing.ID
		} else {
			p.ID = d.newID()
		}
	}
	p.UpdatedAt = d.now()
	if err := d.Logistics.Save(ctx, p); err != nil {
		return domain.LogisticsPolicy{}, err
	}
	return p, nil
}

func (d *Deps) UpsertLegal(ctx context.Context, doc domain.LegalDocument) (domain.LegalDocument, error) {
	if doc.TenantID == uuid.Nil || doc.CountryID == uuid.Nil || doc.Kind == "" || doc.Locale == "" || doc.URI == "" {
		return domain.LegalDocument{}, domain.ErrInvalidArgument
	}
	doc.ID = d.newID()
	doc.Version = 1
	doc.UpdatedAt = d.now()
	if err := d.Legal.Save(ctx, doc); err != nil {
		return domain.LegalDocument{}, err
	}
	return doc, nil
}

func (d *Deps) Resolve(ctx context.Context, tenantID uuid.UUID, countryCode, locale, ns string) (domain.ResolveBundle, error) {
	c, err := d.Countries.GetByCode(ctx, tenantID, strings.ToUpper(countryCode))
	if err != nil {
		return domain.ResolveBundle{}, err
	}
	if c.Status != domain.CountryActive {
		return domain.ResolveBundle{}, domain.ErrCountryInactive
	}
	locale = domain.NormalizeLocale(locale)
	if locale == "" {
		locale = c.DefaultLocale
	}
	loc, err := d.Locales.Get(ctx, tenantID, locale)
	if err != nil {
		loc = domain.LocaleProfile{Locale: locale, LanguageCode: strings.Split(locale, "-")[0], CountryCode: c.Code}
	}
	cur, err := d.Currencies.GetByCode(ctx, tenantID, c.DefaultCurrency)
	if err != nil {
		cur = domain.Currency{Code: c.DefaultCurrency, MinorUnits: 2}
	}
	var fx *domain.ExchangeRate
	if rate, err := d.Rates.Latest(ctx, tenantID, "USD", c.DefaultCurrency); err == nil {
		fx = &rate
	}
	var rules *domain.RegionalRule
	if r, err := d.Rules.GetForCountry(ctx, tenantID, c.ID); err == nil {
		rules = &r
	}
	taxes, _ := d.Taxes.ListByCountry(ctx, tenantID, c.ID)
	var privacy *domain.PrivacyRegime
	if p, err := d.Privacy.GetByCountry(ctx, tenantID, c.ID); err == nil {
		privacy = &p
	}
	pays, _ := d.PayAvail.ListByCountry(ctx, tenantID, c.ID)
	var logistics *domain.LogisticsPolicy
	if l, err := d.Logistics.GetByCountry(ctx, tenantID, c.ID); err == nil {
		logistics = &l
	}
	if ns == "" {
		ns = "app"
	}
	primary, _ := d.Translations.ListNamespace(ctx, tenantID, ns, locale)
	fallbackLocale := c.DefaultLocale
	fallback, _ := d.Translations.ListNamespace(ctx, tenantID, ns, fallbackLocale)
	pm := map[string]string{}
	fm := map[string]string{}
	for _, t := range primary {
		pm[t.Key] = t.Value
	}
	for _, t := range fallback {
		fm[t.Key] = t.Value
	}
	merged := map[string]string{}
	for k := range fm {
		if v, ok := domain.ResolveTranslation(pm, fm, k); ok {
			merged[k] = v
		}
	}
	for k, v := range pm {
		merged[k] = v
	}
	holidays, _ := d.Holidays.ListByCountry(ctx, tenantID, c.ID)
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "global.resolve", map[string]string{"country": c.Code, "locale": locale}, 1)
	}
	return domain.ResolveBundle{
		Country: c, Locale: loc, Currency: cur, FX: fx, Rules: rules, Tax: taxes,
		Privacy: privacy, Payments: pays, Logistics: logistics, Translations: merged, Holidays: holidays,
	}, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	countries, _ := d.Countries.List(ctx, tenantID)
	langs, _ := d.Languages.List(ctx, tenantID)
	curs, _ := d.Currencies.List(ctx, tenantID)
	active := 0
	for _, c := range countries {
		if c.Status == domain.CountryActive {
			active++
		}
	}
	return map[string]any{
		"countries": len(countries), "activeCountries": active,
		"languages": len(langs), "currencies": len(curs),
	}, nil
}

func (d *Deps) SeedTR(ctx context.Context, tenantID uuid.UUID) error {
	c, err := d.UpsertCountry(ctx, domain.Country{
		TenantID: tenantID, Code: "TR", Name: "Türkiye", DefaultLocale: "tr-TR",
		DefaultCurrency: "TRY", DefaultTZ: "Europe/Istanbul", DataResidency: "tr-central", Status: domain.CountryDraft,
	})
	if err != nil {
		return err
	}
	_, _ = d.ActivateCountry(ctx, tenantID, c.ID)
	_, _ = d.AddLanguage(ctx, domain.Language{TenantID: tenantID, Code: "tr", Name: "Türkçe", RTL: false})
	_, _ = d.AddLanguage(ctx, domain.Language{TenantID: tenantID, Code: "en", Name: "English", RTL: false})
	_, _ = d.UpsertLocale(ctx, domain.LocaleProfile{
		TenantID: tenantID, Locale: "tr-TR", LanguageCode: "tr", CountryCode: "TR",
		DateFormat: "DD.MM.YYYY", TimeFormat: "HH:mm", NumberFormat: "1.234,56", CurrencyFormat: "₺#,##0.00", FirstDayOfWeek: 1,
	})
	_, _ = d.UpsertCurrency(ctx, domain.Currency{TenantID: tenantID, Code: "TRY", Name: "Turkish Lira", MinorUnits: 2, Symbol: "₺"})
	_, _ = d.UpsertCurrency(ctx, domain.Currency{TenantID: tenantID, Code: "USD", Name: "US Dollar", MinorUnits: 2, Symbol: "$"})
	_, _ = d.UpsertExchangeRate(ctx, domain.ExchangeRate{TenantID: tenantID, BaseCurrency: "USD", QuoteCurrency: "TRY", Rate: 34.5, Source: "seed"})
	_, _ = d.UpsertRule(ctx, domain.RegionalRule{
		TenantID: tenantID, CountryID: c.ID, MinOrderMinor: 5000, DeliveryFeeMinor: 1499, Currency: "TRY", LegalAge: 18,
		OpenHour: "00:00", CloseHour: "23:59",
	})
	_, _ = d.UpsertTax(ctx, domain.TaxRule{TenantID: tenantID, CountryID: c.ID, Kind: domain.TaxVAT, RateBps: 2000, Name: "KDV %20"})
	_, _ = d.UpsertPrivacy(ctx, domain.PrivacyRegime{
		TenantID: tenantID, CountryID: c.ID, Framework: "KVKK", ConsentRequired: true, RetentionDays: 1095, DataResidency: "tr-central",
	})
	_, _ = d.UpsertPayAvail(ctx, domain.PaymentMethodAvailability{TenantID: tenantID, CountryID: c.ID, MethodCode: "card", Enabled: true})
	_, _ = d.UpsertPayAvail(ctx, domain.PaymentMethodAvailability{TenantID: tenantID, CountryID: c.ID, MethodCode: "wallet", Enabled: true})
	_, _ = d.UpsertLogistics(ctx, domain.LogisticsPolicy{TenantID: tenantID, CountryID: c.ID, SLAMinutes: 30, HolidayRouting: true})
	_, _ = d.UpsertTranslation(ctx, domain.Translation{TenantID: tenantID, Namespace: "app", Key: "home.title", Locale: "tr-TR", Value: "Merhaba"})
	_, _ = d.UpsertTranslation(ctx, domain.Translation{TenantID: tenantID, Namespace: "app", Key: "home.title", Locale: "en", Value: "Hello"})
	_, _ = d.ImportHoliday(ctx, domain.Holiday{TenantID: tenantID, CountryID: c.ID, Date: time.Date(d.now().Year(), 1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02"), Name: "Yılbaşı"})
	_, _ = d.UpsertPlace(ctx, domain.Place{TenantID: tenantID, CountryID: c.ID, Kind: domain.PlaceCity, Code: "IST", Name: "İstanbul"})
	return nil
}
