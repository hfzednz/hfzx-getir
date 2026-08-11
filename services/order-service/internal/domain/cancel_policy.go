package domain

import "fmt"

// CancelPhase classifies where the order sits relative to payment and pick.
type CancelPhase string

const (
	CancelPhaseBeforePayment CancelPhase = "before_payment"
	CancelPhaseAfterPayment  CancelPhase = "after_payment"
	CancelPhaseBeforePick    CancelPhase = "before_pick"
	CancelPhaseDuringPick    CancelPhase = "during_pick"
	CancelPhaseAfterDispatch CancelPhase = "after_dispatch"
	CancelPhaseAfterDelivery CancelPhase = "after_delivery"
	CancelPhaseTerminal      CancelPhase = "terminal"
)

// CancelAction describes compensations required when cancelling.
type CancelAction struct {
	Allowed             bool
	Phase               CancelPhase
	ReleaseReservation  bool
	VoidOrRefundPayment bool
	NotifyWarehouse     bool
	NotifyDispatch      bool
	Reason              string
}

// CancelPhaseFor maps an order status to a cancel policy phase.
func CancelPhaseFor(status OrderStatus) CancelPhase {
	switch status {
	case OrderStatusDraft, OrderStatusPendingPayment:
		return CancelPhaseBeforePayment
	case OrderStatusPaymentProcessing, OrderStatusPaymentFailed:
		return CancelPhaseAfterPayment
	case OrderStatusInventoryReservation, OrderStatusInventoryFailed,
		OrderStatusWarehouseAssigned:
		return CancelPhaseBeforePick
	case OrderStatusPicking, OrderStatusPacking, OrderStatusReadyForDispatch:
		return CancelPhaseDuringPick
	case OrderStatusCourierAssigned, OrderStatusOutForDelivery:
		return CancelPhaseAfterDispatch
	case OrderStatusDelivered, OrderStatusCompleted, OrderStatusRefundPending,
		OrderStatusRefunded:
		return CancelPhaseAfterDelivery
	case OrderStatusCancelled, OrderStatusFailed, OrderStatusArchived:
		return CancelPhaseTerminal
	default:
		return CancelPhaseTerminal
	}
}

// EvaluateCancel returns cancel policy for the given status.
// Rules:
//   - before payment: free cancel, no payment void
//   - after payment / before pick: cancel + release reserve + void/refund
//   - during pick: cancel allowed with WH notify + release + void/refund
//   - after dispatch: cancel blocked (use return after delivery)
//   - after delivery: cancel blocked (use return/refund saga)
//   - terminal: blocked
func EvaluateCancel(status OrderStatus) CancelAction {
	phase := CancelPhaseFor(status)
	switch phase {
	case CancelPhaseBeforePayment:
		return CancelAction{
			Allowed: true,
			Phase:   phase,
			Reason:  "cancel before payment authorization",
		}
	case CancelPhaseAfterPayment:
		return CancelAction{
			Allowed:             true,
			Phase:               phase,
			ReleaseReservation:  status == OrderStatusPaymentProcessing,
			VoidOrRefundPayment: status == OrderStatusPaymentProcessing,
			Reason:              "cancel after payment attempt",
		}
	case CancelPhaseBeforePick:
		return CancelAction{
			Allowed:             true,
			Phase:               phase,
			ReleaseReservation:  true,
			VoidOrRefundPayment: true,
			NotifyWarehouse:     status == OrderStatusWarehouseAssigned,
			Reason:              "cancel before pick",
		}
	case CancelPhaseDuringPick:
		return CancelAction{
			Allowed:             true,
			Phase:               phase,
			ReleaseReservation:  true,
			VoidOrRefundPayment: true,
			NotifyWarehouse:     true,
			Reason:              "cancel during pick/pack (admin/ops)",
		}
	case CancelPhaseAfterDispatch:
		return CancelAction{
			Allowed: false,
			Phase:   phase,
			Reason:  "cancel blocked after dispatch; wait for delivery then return",
		}
	case CancelPhaseAfterDelivery:
		return CancelAction{
			Allowed: false,
			Phase:   phase,
			Reason:  "cancel blocked after delivery; use return/refund saga",
		}
	default:
		return CancelAction{
			Allowed: false,
			Phase:   phase,
			Reason:  "cancel blocked in terminal status",
		}
	}
}

// AssertCancelAllowed returns ErrCancelNotAllowed when cancel is blocked.
func AssertCancelAllowed(status OrderStatus) error {
	action := EvaluateCancel(status)
	if !action.Allowed {
		return fmt.Errorf("%w: %s (%s)", ErrCancelNotAllowed, status, action.Reason)
	}
	return nil
}

// CanModify reports whether pre-pick modifications are allowed.
// Items/address/schedule/notes/gift editable before pick starts.
func CanModify(status OrderStatus) bool {
	switch status {
	case OrderStatusDraft, OrderStatusPendingPayment, OrderStatusPaymentProcessing,
		OrderStatusInventoryReservation, OrderStatusWarehouseAssigned:
		return true
	default:
		return false
	}
}

// AssertModifyAllowed returns ErrModifyNotAllowed when modifications are blocked.
func AssertModifyAllowed(status OrderStatus) error {
	if !CanModify(status) {
		return fmt.Errorf("%w: status %s", ErrModifyNotAllowed, status)
	}
	return nil
}
