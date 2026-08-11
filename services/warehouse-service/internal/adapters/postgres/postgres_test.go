package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

func TestOpenRequiresURL(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("expected error for empty DATABASE_URL")
	}
}

func TestJSONMapRoundTrip(t *testing.T) {
	m := JSONMap{"a": "b", "n": float64(1)}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out JSONMap
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if out["a"] != "b" {
		t.Fatalf("got %#v", out)
	}
}

func TestTextArrayJSONRoundTrip(t *testing.T) {
	a := TextArray{"box", "tape"}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out TextArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != "box" {
		t.Fatalf("got %#v", out)
	}
}

func TestNullHelpers(t *testing.T) {
	id := uuid.New()
	if nullUUID(nil) != nil {
		t.Fatal("nil pointer")
	}
	if nullUUID(&id) != id {
		t.Fatal("uuid value")
	}
	if nullUUIDValue(uuid.Nil) != nil {
		t.Fatal("nil uuid")
	}
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&time.Time{}) != nil {
		t.Fatal("null time")
	}
	if nullTime(&now) != now {
		t.Fatal("time value")
	}
	n := uuid.NullUUID{UUID: id, Valid: true}
	if got := scanNullUUID(n); got == nil || *got != id {
		t.Fatal("scanNullUUID")
	}
	if scanUUIDOrNil(uuid.NullUUID{}) != uuid.Nil {
		t.Fatal("scanUUIDOrNil")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.FulfillmentRepo = (*FulfillmentRepo)(nil)
		_ ports.TaskRepo        = (*TaskRepo)(nil)
		_ ports.PickRepo        = (*PickRepo)(nil)
		_ ports.PackRepo        = (*PackRepo)(nil)
		_ ports.DispatchRepo    = (*DispatchRepo)(nil)
		_ ports.StationRepo     = (*StationRepo)(nil)
		_ ports.WorkforceRepo   = (*WorkforceRepo)(nil)
		_ ports.EquipmentRepo   = (*EquipmentRepo)(nil)
		_ ports.QCRepo          = (*QCRepo)(nil)
		_ ports.LabelRepo       = (*LabelRepo)(nil)
	)
	_ = domain.FulfillmentStatusReceived
	b, _ := json.Marshal(domain.PickRouteStep{LineID: uuid.New(), Seq: 1})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Fulfillments == nil || repos.Labels == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
