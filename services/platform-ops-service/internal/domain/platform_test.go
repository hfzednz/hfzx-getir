package domain_test

import (
	"testing"

	"github.com/nexora/platform-ops-service/internal/domain"
)

func TestBurnAndStrategy(t *testing.T) {
	if !domain.ValidStrategy("canary") {
		t.Fatal("canary")
	}
	b := domain.BurnRate(99.9, 99.8)
	if b <= 0 {
		t.Fatalf("burn %v", b)
	}
}
