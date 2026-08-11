package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SagaType classifies orchestration workflows owned by OMS.
type SagaType string

const (
	SagaTypePlace  SagaType = "place"
	SagaTypeCancel SagaType = "cancel"
	SagaTypeRefund SagaType = "refund"
	SagaTypeReturn SagaType = "return"
)

// Valid reports whether the saga type is recognized.
func (t SagaType) Valid() bool {
	switch t {
	case SagaTypePlace, SagaTypeCancel, SagaTypeRefund, SagaTypeReturn:
		return true
	default:
		return false
	}
}

// SagaInstanceStatus is the lifecycle of a saga header.
type SagaInstanceStatus string

const (
	SagaInstancePending      SagaInstanceStatus = "pending"
	SagaInstanceRunning      SagaInstanceStatus = "running"
	SagaInstanceCompensating SagaInstanceStatus = "compensating"
	SagaInstanceSucceeded    SagaInstanceStatus = "succeeded"
	SagaInstanceFailed       SagaInstanceStatus = "failed"
	SagaInstanceCompensated  SagaInstanceStatus = "compensated"
)

// Valid reports whether the saga instance status is recognized.
func (s SagaInstanceStatus) Valid() bool {
	switch s {
	case SagaInstancePending, SagaInstanceRunning, SagaInstanceCompensating,
		SagaInstanceSucceeded, SagaInstanceFailed, SagaInstanceCompensated:
		return true
	default:
		return false
	}
}

// SagaStepStatus is the lifecycle of an individual saga step.
type SagaStepStatus string

const (
	SagaStepPending     SagaStepStatus = "pending"
	SagaStepSucceeded   SagaStepStatus = "succeeded"
	SagaStepFailed      SagaStepStatus = "failed"
	SagaStepCompensated SagaStepStatus = "compensated"
)

// Valid reports whether the saga step status is recognized.
func (s SagaStepStatus) Valid() bool {
	switch s {
	case SagaStepPending, SagaStepSucceeded, SagaStepFailed, SagaStepCompensated:
		return true
	default:
		return false
	}
}

// Place-saga forward steps (happy path).
const (
	SagaStepValidate         = "Validate"
	SagaStepSoftReserve      = "SoftReserve"
	SagaStepAuthorizePayment = "AuthorizePayment"
	SagaStepConfirmHard      = "ConfirmHard"
	SagaStepStartFulfillment = "StartFulfillment"
	SagaStepRequestDispatch  = "RequestDispatch"
	SagaStepComplete         = "Complete"
)

// Place-saga compensations.
const (
	SagaStepReleaseReserve      = "ReleaseReserve"
	SagaStepVoidOrRefundPayment = "VoidOrRefundPayment"
)

// PlaceSagaForwardSteps is the ordered happy-path place saga.
var PlaceSagaForwardSteps = []string{
	SagaStepValidate,
	SagaStepSoftReserve,
	SagaStepAuthorizePayment,
	SagaStepConfirmHard,
	SagaStepStartFulfillment,
	SagaStepRequestDispatch,
	SagaStepComplete,
}

// CompensationFor maps a forward place step to its compensation (if any).
func CompensationFor(forwardStep string) string {
	switch forwardStep {
	case SagaStepSoftReserve, SagaStepConfirmHard:
		return SagaStepReleaseReserve
	case SagaStepAuthorizePayment:
		return SagaStepVoidOrRefundPayment
	default:
		return ""
	}
}

// IsKnownSagaStep reports whether name is a known place-saga step or compensation.
func IsKnownSagaStep(name string) bool {
	switch name {
	case SagaStepValidate, SagaStepSoftReserve, SagaStepAuthorizePayment,
		SagaStepConfirmHard, SagaStepStartFulfillment, SagaStepRequestDispatch,
		SagaStepComplete, SagaStepReleaseReserve, SagaStepVoidOrRefundPayment:
		return true
	default:
		return false
	}
}

// SagaInstance is an orchestration header for place|cancel|refund|return.
type SagaInstance struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	TenantID       uuid.UUID
	SagaType       SagaType
	Status         SagaInstanceStatus
	CurrentStep    string
	CorrelationID  string
	IdempotencyKey string
	LastError      string
	Steps          []SagaStep
	Metadata       map[string]any
	StartedAt      *time.Time
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SagaStep is a single forward or compensation step with retry metadata.
type SagaStep struct {
	ID             uuid.UUID
	SagaID         uuid.UUID
	OrderID        uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Status         SagaStepStatus
	Attempt        int
	LastError      string
	IdempotencyKey string
	CompensationOf string
	Payload        map[string]any
	StartedAt      *time.Time
	CompletedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Validate checks saga instance invariants.
func (s SagaInstance) Validate() error {
	if s.ID == uuid.Nil {
		return fmt.Errorf("%w: saga id required", ErrInvalidArgument)
	}
	if s.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if s.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if !s.SagaType.Valid() {
		return fmt.Errorf("%w: invalid saga_type %q", ErrInvalidArgument, s.SagaType)
	}
	if !s.Status.Valid() {
		return fmt.Errorf("%w: invalid saga status %q", ErrInvalidArgument, s.Status)
	}
	if s.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	for i, step := range s.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("step[%d]: %w", i, err)
		}
	}
	return nil
}

// Validate checks saga step invariants.
func (st SagaStep) Validate() error {
	if st.ID == uuid.Nil {
		return fmt.Errorf("%w: saga step id required", ErrInvalidArgument)
	}
	if st.SagaID == uuid.Nil {
		return fmt.Errorf("%w: saga_id required", ErrInvalidArgument)
	}
	if st.OrderID == uuid.Nil {
		return fmt.Errorf("%w: order_id required", ErrInvalidArgument)
	}
	if st.TenantID == uuid.Nil {
		return fmt.Errorf("%w: tenant_id required", ErrInvalidArgument)
	}
	if st.Name == "" {
		return fmt.Errorf("%w: step name required", ErrInvalidArgument)
	}
	if !st.Status.Valid() {
		return fmt.Errorf("%w: invalid step status %q", ErrInvalidArgument, st.Status)
	}
	if st.IdempotencyKey == "" {
		return fmt.Errorf("%w: idempotency_key required", ErrInvalidArgument)
	}
	if st.Attempt < 0 {
		return fmt.Errorf("%w: attempt must be >= 0", ErrInvalidArgument)
	}
	return nil
}

// NewPlaceSagaSteps builds pending forward steps for a place saga.
func NewPlaceSagaSteps(sagaID, orderID, tenantID uuid.UUID, keyPrefix string) []SagaStep {
	now := time.Now().UTC()
	steps := make([]SagaStep, 0, len(PlaceSagaForwardSteps))
	for _, name := range PlaceSagaForwardSteps {
		steps = append(steps, SagaStep{
			ID:             uuid.New(),
			SagaID:         sagaID,
			OrderID:        orderID,
			TenantID:       tenantID,
			Name:           name,
			Status:         SagaStepPending,
			Attempt:        0,
			IdempotencyKey: fmt.Sprintf("%s:%s", keyPrefix, name),
			CreatedAt:      now,
			UpdatedAt:      now,
		})
	}
	return steps
}
