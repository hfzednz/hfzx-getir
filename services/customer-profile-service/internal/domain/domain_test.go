package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
)

func TestAddress_ValidateCoords(t *testing.T) {
	a := domain.Address{
		ID:        uuid.New(),
		ProfileID: uuid.New(),
		TenantID:  uuid.New(),
		Label:     domain.AddressLabelHome,
		Line1:     "1 Main",
		Lat:       41.0,
		Lng:       29.0,
	}
	if err := a.Validate(); err != nil {
		t.Fatalf("valid address: %v", err)
	}
	a.Lat = 120
	if err := a.Validate(); err == nil {
		t.Fatal("expected invalid lat")
	}
}

func TestConsentChannel_Valid(t *testing.T) {
	if !domain.ConsentChannelMarketing.Valid() {
		t.Fatal("marketing should be valid")
	}
}
