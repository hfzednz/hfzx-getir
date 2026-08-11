package postgres

import (
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/app/ports"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"k": "v"}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out["k"] != "v" {
		t.Fatalf("got %#v", out)
	}
}

func TestUUIDArrayRoundTrip(t *testing.T) {
	id := uuid.New()
	a := UUIDArray{id}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out UUIDArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] != id {
		t.Fatalf("got %#v", out)
	}
}

func TestIntArrayRoundTrip(t *testing.T) {
	a := IntArray{1, 2, 23}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out IntArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 3 || out[2] != 23 {
		t.Fatalf("got %#v", out)
	}
}

func TestNullHelpers(t *testing.T) {
	id := uuid.New()
	if nullUUID(nil) != nil || nullUUIDValue(uuid.Nil) != nil {
		t.Fatal("null uuid")
	}
	if nullUUID(&id) != id {
		t.Fatal("uuid")
	}
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&now) != now {
		t.Fatal("time")
	}
	ip := net.ParseIP("127.0.0.1")
	if nullIP(nil) != nil || nullIP(ip) != "127.0.0.1" {
		t.Fatal("ip")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.ProfileRepository         = (*ProfileRepo)(nil)
		_ ports.AddressRepository         = (*AddressRepo)(nil)
		_ ports.PreferencesRepository     = (*PreferencesRepo)(nil)
		_ ports.TagRepository             = (*TagRepo)(nil)
		_ ports.HouseholdRepository       = (*HouseholdRepo)(nil)
		_ ports.ConsentRepository         = (*ConsentRepo)(nil)
		_ ports.CRMRepository             = (*CRMRepo)(nil)
		_ ports.SegmentRepository         = (*SegmentRepo)(nil)
		_ ports.PersonalizationRepository = (*PersonalizationRepo)(nil)
		_ ports.AIModelRepository         = (*AIModelRepo)(nil)
		_ ports.PrivacyRepository         = (*PrivacyRepo)(nil)
		_ ports.ActivityRepository        = (*ActivityRepo)(nil)
	)
	repos := NewRepos(nil)
	if repos.Profiles == nil || repos.Activity == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
