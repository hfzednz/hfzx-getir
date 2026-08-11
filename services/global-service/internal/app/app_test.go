package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/global-service/internal/app"
	"github.com/nexora/global-service/internal/app/memory"
	"github.com/nexora/global-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Countries: r.Countries, Places: r.Places, Languages: r.Languages, Locales: r.Locales,
		Translations: r.Translations, Currencies: r.Currencies, Rates: r.Rates, Holidays: r.Holidays,
		Rules: r.Rules, Taxes: r.Taxes, Privacy: r.Privacy, PayAvail: r.PayAvail, Logistics: r.Logistics,
		Legal: r.Legal, Outbox: r.Outbox, FX: r.FX, AI: r.AI, Metrics: r.Metrics, Cache: r.Cache,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestGlobalResolveAndFX(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.SeedTR(ctx, tid); err != nil {
		t.Fatal(err)
	}

	bundle, err := d.Resolve(ctx, tid, "TR", "tr-TR", "app")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Country.Code != "TR" || bundle.Translations["home.title"] != "Merhaba" {
		t.Fatalf("%+v", bundle.Translations)
	}
	if bundle.Privacy == nil || bundle.Privacy.Framework != "KVKK" {
		t.Fatal(bundle.Privacy)
	}
	if len(bundle.Tax) == 0 || bundle.Tax[0].RateBps != 2000 {
		t.Fatal(bundle.Tax)
	}

	out, err := d.Convert(ctx, tid, 10000, "USD", "TRY")
	if err != nil || out <= 0 {
		t.Fatalf("%d %v", out, err)
	}

	ai, err := d.AIAssistTranslate(ctx, tid, "app", "cart.title", "en", "tr-TR", "Cart")
	if err != nil || ai.Value == "" {
		t.Fatal(err)
	}

	tax := domain.TaxAmountMinor(10000, 2000)
	if tax != 2000 {
		t.Fatal(tax)
	}

	st, _ := d.AdminStats(ctx, tid)
	if st["activeCountries"].(int) < 1 {
		t.Fatal(st)
	}
}
