package memory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/nexora/order-service/internal/app/ports"
)

// ErrInjectedFailure is returned when a client failure flag is set.
var ErrInjectedFailure = errors.New("injected client failure")

// InventoryClient is a memory inventory client that succeeds by default.
type InventoryClient struct {
	mu sync.Mutex

	FailSoftReserve bool
	FailConfirmHard bool
	FailRelease     bool

	SoftReserveCalls atomic.Int64
	ConfirmHardCalls atomic.Int64
	ReleaseCalls     atomic.Int64

	// SeenSoftKeys tracks soft-reserve idempotency keys (for double-reserve tests).
	SeenSoftKeys map[string]string
}

// NewInventoryClient returns a succeeding inventory client.
func NewInventoryClient() *InventoryClient {
	return &InventoryClient{SeenSoftKeys: make(map[string]string)}
}

func (c *InventoryClient) SoftReserve(_ context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	c.SoftReserveCalls.Add(1)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.FailSoftReserve {
		return ports.SoftReserveResult{}, fmt.Errorf("%w: soft reserve", ErrInjectedFailure)
	}
	if c.SeenSoftKeys == nil {
		c.SeenSoftKeys = make(map[string]string)
	}
	if ref, ok := c.SeenSoftKeys[req.IdempotencyKey]; ok {
		return ports.SoftReserveResult{ReservationRef: ref}, nil
	}
	ref := "res-" + uuid.NewString()
	c.SeenSoftKeys[req.IdempotencyKey] = ref
	return ports.SoftReserveResult{ReservationRef: ref}, nil
}

func (c *InventoryClient) ConfirmHard(_ context.Context, _ ports.ConfirmHardRequest) error {
	c.ConfirmHardCalls.Add(1)
	if c.FailConfirmHard {
		return fmt.Errorf("%w: confirm hard", ErrInjectedFailure)
	}
	return nil
}

func (c *InventoryClient) Release(_ context.Context, _ ports.ReleaseRequest) error {
	c.ReleaseCalls.Add(1)
	if c.FailRelease {
		return fmt.Errorf("%w: release", ErrInjectedFailure)
	}
	return nil
}

var _ ports.InventoryClient = (*InventoryClient)(nil)

// PaymentClient is a memory payment client that succeeds by default.
type PaymentClient struct {
	FailAuthorize bool
	FailVoid      bool
	FailRefund    bool

	AuthorizeCalls atomic.Int64
	VoidCalls      atomic.Int64
	RefundCalls    atomic.Int64
}

func (c *PaymentClient) Authorize(_ context.Context, _ ports.AuthorizeRequest) (ports.AuthorizeResult, error) {
	c.AuthorizeCalls.Add(1)
	if c.FailAuthorize {
		return ports.AuthorizeResult{}, fmt.Errorf("%w: authorize", ErrInjectedFailure)
	}
	return ports.AuthorizeResult{PaymentIntentRef: "pi-" + uuid.NewString()}, nil
}

func (c *PaymentClient) Void(_ context.Context, _ ports.VoidRequest) error {
	c.VoidCalls.Add(1)
	if c.FailVoid {
		return fmt.Errorf("%w: void", ErrInjectedFailure)
	}
	return nil
}

func (c *PaymentClient) Refund(_ context.Context, _ ports.RefundPaymentRequest) (ports.RefundPaymentResult, error) {
	c.RefundCalls.Add(1)
	if c.FailRefund {
		return ports.RefundPaymentResult{}, fmt.Errorf("%w: refund", ErrInjectedFailure)
	}
	return ports.RefundPaymentResult{PaymentRefundRef: "rf-" + uuid.NewString()}, nil
}

var _ ports.PaymentClient = (*PaymentClient)(nil)

// WarehouseClient is a memory warehouse client that succeeds by default.
type WarehouseClient struct {
	FailReceive      bool
	ReceiveCalls     atomic.Int64
	LastFulfillment  string
}

func (c *WarehouseClient) ReceiveFulfillment(_ context.Context, _ ports.ReceiveFulfillmentRequest) (ports.ReceiveFulfillmentResult, error) {
	c.ReceiveCalls.Add(1)
	if c.FailReceive {
		return ports.ReceiveFulfillmentResult{}, fmt.Errorf("%w: receive fulfillment", ErrInjectedFailure)
	}
	ref := "ful-" + uuid.NewString()
	c.LastFulfillment = ref
	return ports.ReceiveFulfillmentResult{FulfillmentRef: ref}, nil
}

var _ ports.WarehouseClient = (*WarehouseClient)(nil)

// DispatchClient is a memory dispatch client that succeeds by default.
type DispatchClient struct {
	FailRequest  bool
	RequestCalls atomic.Int64
}

func (c *DispatchClient) RequestDispatch(_ context.Context, _ ports.RequestDispatchRequest) (ports.RequestDispatchResult, error) {
	c.RequestCalls.Add(1)
	if c.FailRequest {
		return ports.RequestDispatchResult{}, fmt.Errorf("%w: request dispatch", ErrInjectedFailure)
	}
	return ports.RequestDispatchResult{DispatchRef: "dsp-" + uuid.NewString()}, nil
}

var _ ports.DispatchClient = (*DispatchClient)(nil)
