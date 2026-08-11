package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/review-service/internal/app/ports"
	"github.com/nexora/review-service/internal/domain"
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

func TestTextArrayRoundTrip(t *testing.T) {
	a := TextArray{"fresh", "packaging"}
	v, err := a.Value()
	if err != nil {
		t.Fatal(err)
	}
	var out TextArray
	if err := out.Scan(v); err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0] != "fresh" {
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
	now := time.Now().UTC()
	if nullTime(nil) != nil || nullTime(&time.Time{}) != nil {
		t.Fatal("null time")
	}
	if nullTime(&now) != now {
		t.Fatal("time value")
	}
}

func TestInterfaceCompliance(t *testing.T) {
	var (
		_ ports.ReviewRepo       = (*ReviewRepo)(nil)
		_ ports.RatingRepo       = (*RatingRepo)(nil)
		_ ports.MediaRepo        = (*MediaRepo)(nil)
		_ ports.InteractionRepo  = (*InteractionRepo)(nil)
		_ ports.QualityRepo      = (*QualityRepo)(nil)
		_ ports.ModerationRepo   = (*ModerationRepo)(nil)
		_ ports.TrustRepo        = (*TrustRepo)(nil)
		_ ports.ReputationRepo   = (*ReputationRepo)(nil)
		_ ports.OutboxRepository = (*OutboxRepo)(nil)
	)
	_ = domain.ReviewStatusPublished
	b, _ := json.Marshal(domain.TrustScore{Score: 40})
	if len(b) == 0 {
		t.Fatal("marshal")
	}
	repos := NewRepos(nil)
	if repos.Reviews == nil || repos.Outbox == nil {
		t.Fatal("NewRepos")
	}
}

func TestMapUniqueViolationNil(t *testing.T) {
	if mapUniqueViolation(nil) != nil {
		t.Fatal("nil")
	}
}
