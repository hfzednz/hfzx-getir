package payout

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/settlement-service/internal/app/ports"
)

func TestExecuteUnavailableWithoutProviderURL(t *testing.T) {
	c := NewClient("", nil)
	res, err := c.Execute(context.Background(), ports.PayoutRequest{
		InstructionID: uuid.New(),
		TenantID:      uuid.New(),
		PayeeType:     "courier",
		PayeeRef:      "c1",
		AmountMinor:   1000,
		Currency:      "TRY",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Succeeded {
		t.Fatal("empty PAYOUT_PROVIDER_URL must not fabricate a paid payout")
	}
	if res.Error != "provider_unavailable" {
		t.Fatalf("error=%q", res.Error)
	}
}
