package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app"
	"github.com/nexora/customer-profile-service/internal/app/memory"
	"github.com/nexora/customer-profile-service/internal/app/ports"
	"github.com/nexora/customer-profile-service/internal/domain"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store, *memory.EventPublisher) {
	t.Helper()
	store := memory.NewStore()
	events := &memory.EventPublisher{}
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	d := &app.Deps{
		Profiles:        &memory.ProfileRepo{S: store},
		Addresses:       &memory.AddressRepo{S: store},
		Preferences:     &memory.PreferencesRepo{S: store},
		Tags:            &memory.TagRepo{S: store},
		Households:      &memory.HouseholdRepo{S: store},
		Consents:        &memory.ConsentRepo{S: store},
		CRM:             &memory.CRMRepo{S: store},
		Segments:        &memory.SegmentRepo{S: store},
		Personalization: &memory.PersonalizationRepo{S: store},
		AIModels:        &memory.AIModelRepo{S: store},
		Privacy:         &memory.PrivacyRepo{S: store},
		Activity:        &memory.ActivityRepo{S: store},
		Events:          events,
		Media:           memory.NewMediaStore(),
		Zones:           &memory.ZoneValidator{OK: true, ZoneID: "zone-test"},
		Clock:           clock,
		IDs:             memory.IDGen{},
	}
	return d, store, events
}

func TestProvisionAndUpdate(t *testing.T) {
	d, _, events := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	principal := uuid.New()

	p, err := d.Provision(ctx, app.ProvisionInput{
		TenantID:    tenant,
		PrincipalID: principal,
		DisplayName: "Ada Lovelace",
		FullName:    "Augusta Ada King",
		Language:    "tr-TR",
		CountryCode: "TR",
	})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}
	if p.ID == uuid.Nil || p.Status != domain.ProfileStatusActive {
		t.Fatalf("unexpected profile: %+v", p)
	}
	created := events.OfType(domain.EventCustomerCreated)
	if len(created) != 1 {
		t.Fatalf("expected CustomerCreated event, got %d", len(created))
	}
	if created[0].Topic != ports.TopicProfileLifecycle {
		t.Fatalf("topic = %s", created[0].Topic)
	}

	p2, err := d.Provision(ctx, app.ProvisionInput{TenantID: tenant, PrincipalID: principal})
	if err != nil {
		t.Fatalf("Provision idempotent: %v", err)
	}
	if p2.ID != p.ID {
		t.Fatalf("expected same profile id")
	}

	name := "Ada L."
	updated, err := d.UpdateProfile(ctx, app.UpdateProfileInput{
		ProfileID:   p.ID,
		DisplayName: &name,
		Dietary:     map[string]any{"vegetarian": true},
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.DisplayName != "Ada L." || updated.Dietary["vegetarian"] != true {
		t.Fatalf("unexpected update: %+v", updated)
	}
	if len(events.OfType(domain.EventCustomerUpdated)) != 1 {
		t.Fatalf("expected CustomerUpdated event")
	}
}

func TestAddressDefault(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	p, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(), DisplayName: "Addr User",
	})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	a1, err := d.AddAddress(ctx, app.AddAddressInput{
		ProfileID: p.ID, Label: domain.AddressLabelHome, Line1: "Street 1",
		Lat: 41.0, Lng: 29.0,
	})
	if err != nil {
		t.Fatalf("AddAddress1: %v", err)
	}
	if !a1.IsDefault {
		t.Fatal("first address should be default")
	}
	if !a1.IsZoneValidated() {
		t.Fatalf("expected zone validation: %+v", a1)
	}

	a2, err := d.AddAddress(ctx, app.AddAddressInput{
		ProfileID: p.ID, Label: domain.AddressLabelWork, Line1: "Street 2",
		IsDefault: true, Lat: 41.1, Lng: 29.1,
	})
	if err != nil {
		t.Fatalf("AddAddress2: %v", err)
	}
	if !a2.IsDefault {
		t.Fatal("second address should be default")
	}

	got1, err := d.Addresses.GetByID(ctx, a1.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got1.IsDefault {
		t.Fatal("first address should no longer be default")
	}

	set, err := d.SetDefaultAddress(ctx, p.ID, a1.ID, "")
	if err != nil {
		t.Fatalf("SetDefaultAddress: %v", err)
	}
	if !set.IsDefault {
		t.Fatal("expected a1 default after SetDefault")
	}
	got2, _ := d.Addresses.GetByID(ctx, a2.ID)
	if got2.IsDefault {
		t.Fatal("a2 should not be default")
	}
}

func TestConsentChangeEvent(t *testing.T) {
	d, _, events := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	p, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	c, err := d.SetConsent(ctx, app.SetConsentInput{
		ProfileID: p.ID,
		Channel:   domain.ConsentChannelMarketing,
		Granted:   true,
		Source:    "app",
	})
	if err != nil {
		t.Fatalf("SetConsent: %v", err)
	}
	if !c.Granted {
		t.Fatalf("expected granted consent: %+v", c)
	}

	changed := events.OfType(domain.EventConsentChanged)
	if len(changed) != 1 {
		t.Fatalf("expected ConsentChanged, got %d", len(changed))
	}
	if changed[0].Topic != ports.TopicConsentEvents {
		t.Fatalf("topic = %s", changed[0].Topic)
	}
	payload := changed[0].Payload.(map[string]any)
	if payload["channel"] != string(domain.ConsentChannelMarketing) || payload["granted"] != true {
		t.Fatalf("unexpected payload: %+v", payload)
	}

	_, err = d.SetConsent(ctx, app.SetConsentInput{
		ProfileID: p.ID, Channel: domain.ConsentChannelMarketing, Granted: false,
	})
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if len(events.OfType(domain.EventConsentChanged)) != 2 {
		t.Fatal("expected second ConsentChanged on revoke")
	}
}

func TestMergeSoftDeleteSource(t *testing.T) {
	d, store, events := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	target, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(), DisplayName: "Keep",
	})
	if err != nil {
		t.Fatalf("target: %v", err)
	}
	source, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(), DisplayName: "Merge",
	})
	if err != nil {
		t.Fatalf("source: %v", err)
	}
	_, _ = d.AddAddress(ctx, app.AddAddressInput{
		ProfileID: source.ID, Label: domain.AddressLabelHome, Line1: "Src St",
	})
	_, _ = d.AddTag(ctx, app.AddTagInput{ProfileID: source.ID, Kind: domain.TagKindVIP, Name: "vip"})

	merged, err := d.MergeCustomers(ctx, target.ID, source.ID, "trace-merge")
	if err != nil {
		t.Fatalf("MergeCustomers: %v", err)
	}
	if merged.ID != target.ID {
		t.Fatal("expected target returned")
	}

	src, err := d.Profiles.GetByID(ctx, source.ID)
	if err != nil {
		t.Fatalf("get source: %v", err)
	}
	if src.Status != domain.ProfileStatusMerged {
		t.Fatalf("source status = %s, want merged", src.Status)
	}
	if src.IsActive() {
		t.Fatal("source should not be active")
	}
	if store.Profiles[source.ID].Status != domain.ProfileStatusMerged {
		t.Fatal("store should reflect merged source")
	}

	deletedEvts := events.OfType(domain.EventProfileDeleted)
	found := false
	for _, ev := range deletedEvts {
		m := ev.Payload.(map[string]any)
		if m["reason"] == "merge" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected ProfileDeleted with merge reason")
	}

	tags, _ := d.Tags.List(ctx, target.ID)
	if len(tags) == 0 {
		t.Fatal("expected tags moved to target")
	}
}

func TestPersonalizationUpsert(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	p, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	empty, err := d.GetPersonalization(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetPersonalization empty: %v", err)
	}
	if empty.ProfileID != p.ID {
		t.Fatal("expected default personalization")
	}

	pers, err := d.UpdatePersonalization(ctx, app.UpdatePersonalizationInput{
		ProfileID: p.ID,
		Homepage:  map[string]any{"rail": "fresh"},
		Category:  map[string]any{"coffee": 0.9},
		Search:    map[string]any{"home_rail": "B"},
	})
	if err != nil {
		t.Fatalf("UpdatePersonalization: %v", err)
	}
	if pers.Homepage["rail"] != "fresh" || pers.Category["coffee"] != 0.9 {
		t.Fatalf("unexpected pers: %+v", pers)
	}

	got, err := d.GetPersonalization(ctx, p.ID)
	if err != nil {
		t.Fatalf("GetPersonalization: %v", err)
	}
	if got.Search["home_rail"] != "B" {
		t.Fatalf("search prefs not persisted: %+v", got)
	}

	pers2, err := d.UpdatePersonalization(ctx, app.UpdatePersonalizationInput{
		ProfileID: p.ID,
		Homepage:  map[string]any{"rail": "deals"},
	})
	if err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	if pers2.Homepage["rail"] != "deals" {
		t.Fatal("expected replaced homepage")
	}
}

func TestCheckConsent(t *testing.T) {
	d, _, _ := testDeps(t)
	ctx := context.Background()
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	p, err := d.Provision(ctx, app.ProvisionInput{
		TenantID: tenant, PrincipalID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Provision: %v", err)
	}

	ok, err := d.CheckConsent(ctx, p.ID, domain.ConsentChannelMarketing)
	if err != nil {
		t.Fatalf("CheckConsent missing: %v", err)
	}
	if ok {
		t.Fatal("expected false when no consent recorded")
	}

	_, err = d.SetConsent(ctx, app.SetConsentInput{
		ProfileID: p.ID, Channel: domain.ConsentChannelMarketing, Granted: true,
	})
	if err != nil {
		t.Fatalf("SetConsent: %v", err)
	}
	ok, err = d.CheckConsent(ctx, p.ID, domain.ConsentChannelMarketing)
	if err != nil || !ok {
		t.Fatalf("expected granted=true, got %v err=%v", ok, err)
	}

	_, err = d.SetConsent(ctx, app.SetConsentInput{
		ProfileID: p.ID, Channel: domain.ConsentChannelMarketing, Granted: false,
	})
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	ok, err = d.CheckConsent(ctx, p.ID, domain.ConsentChannelMarketing)
	if err != nil || ok {
		t.Fatalf("expected granted=false after revoke, got %v", ok)
	}

	_, err = d.CheckConsent(ctx, uuid.Nil, domain.ConsentChannelMarketing)
	if err == nil || !strings.Contains(err.Error(), "profile_id") {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}
