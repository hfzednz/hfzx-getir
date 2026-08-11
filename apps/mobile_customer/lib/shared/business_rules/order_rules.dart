import 'package:nexora_core/nexora_core.dart';

/// Order lifecycle states aligned with backend order state machine.
enum OrderLifecycleStatus {
  draft,
  pendingPayment,
  confirmed,
  picking,
  inPicking,
  packed,
  assigned,
  dispatched,
  pickedUp,
  delivered,
  cancelled,
  refunded,
  refundPartial,
  partiallyCancelled,
}

OrderLifecycleStatus orderLifecycleStatusFromJson(String? raw) {
  switch (raw?.toLowerCase()) {
    case 'draft':
      return OrderLifecycleStatus.draft;
    case 'pending_payment':
    case 'pendingpayment':
      return OrderLifecycleStatus.pendingPayment;
    case 'confirmed':
      return OrderLifecycleStatus.confirmed;
    case 'picking':
    case 'in_picking':
      return OrderLifecycleStatus.inPicking;
    case 'packed':
      return OrderLifecycleStatus.packed;
    case 'assigned':
      return OrderLifecycleStatus.assigned;
    case 'dispatched':
    case 'out_for_delivery':
      return OrderLifecycleStatus.dispatched;
    case 'picked_up':
    case 'pickedup':
      return OrderLifecycleStatus.pickedUp;
    case 'delivered':
      return OrderLifecycleStatus.delivered;
    case 'cancelled':
    case 'canceled':
      return OrderLifecycleStatus.cancelled;
    case 'refunded':
      return OrderLifecycleStatus.refunded;
    case 'refund_partial':
    case 'partial_refund':
      return OrderLifecycleStatus.refundPartial;
    case 'partially_cancelled':
    case 'partially_canceled':
      return OrderLifecycleStatus.partiallyCancelled;
    default:
      return OrderLifecycleStatus.confirmed;
  }
}

enum OrderRuleViolationCode {
  notCancellable,
  partialCancelNotAllowed,
  reorderNotEligible,
}

class OrderRuleViolation {
  const OrderRuleViolation({required this.code, required this.message});

  final OrderRuleViolationCode code;
  final String message;
}

/// Order cancellation, partial cancel, and reorder eligibility rules.
abstract final class OrderRules {
  static const _cancellableStatuses = {
    OrderLifecycleStatus.draft,
    OrderLifecycleStatus.pendingPayment,
    OrderLifecycleStatus.confirmed,
  };

  static const _partialCancelStatuses = {
    OrderLifecycleStatus.confirmed,
    OrderLifecycleStatus.picking,
    OrderLifecycleStatus.inPicking,
  };

  static const _reorderEligibleStatuses = {
    OrderLifecycleStatus.delivered,
    OrderLifecycleStatus.cancelled,
    OrderLifecycleStatus.partiallyCancelled,
    OrderLifecycleStatus.refunded,
    OrderLifecycleStatus.refundPartial,
  };

  static bool isCancellable(OrderLifecycleStatus status) =>
      _cancellableStatuses.contains(status);

  static bool allowsPartialCancel(OrderLifecycleStatus status) =>
      _partialCancelStatuses.contains(status);

  static bool isReorderEligible(OrderLifecycleStatus status) =>
      _reorderEligibleStatuses.contains(status);

  static Result<void> validateCancel({
    required OrderLifecycleStatus status,
    required bool paymentCaptured,
  }) {
    if (!isCancellable(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'This order can no longer be cancelled',
          details: {'status': status.name},
        ),
      );
    }

    if (status == OrderLifecycleStatus.pendingPayment && paymentCaptured) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Payment is processing — try again shortly',
          details: {'status': status.name, 'payment_captured': true},
        ),
      );
    }

    return const Success(null);
  }

  static Result<void> validatePartialCancel({
    required OrderLifecycleStatus status,
    required List<String> lineIdsToCancel,
    required int totalLineCount,
  }) {
    if (lineIdsToCancel.isEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Select at least one item to cancel',
          details: {'field': 'line_ids'},
        ),
      );
    }

    if (lineIdsToCancel.length >= totalLineCount) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Use full order cancellation instead',
          details: {'field': 'line_ids'},
        ),
      );
    }

    if (!allowsPartialCancel(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Partial cancellation is not available for this order',
          details: {'status': status.name},
        ),
      );
    }

    return const Success(null);
  }

  static Result<void> validateReorder({
    required OrderLifecycleStatus status,
    required bool allItemsAvailable,
  }) {
    if (!isReorderEligible(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Reorder is not available for this order yet',
          details: {'status': status.name},
        ),
      );
    }

    if (!allItemsAvailable) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Some items from this order are no longer available',
          details: {'status': status.name},
        ),
      );
    }

    return const Success(null);
  }
}
