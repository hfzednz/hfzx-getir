import 'package:nexora_core/nexora_core.dart';

/// Inventory adjustment reasons + FEFO helpers.
abstract final class InventoryRules {
  static const adjustmentReasons = <String>{
    'damage',
    'expiry',
    'theft',
    'found',
    'cycle_count',
    'receive_correction',
    'transfer_loss',
    'other',
  };

  static bool isValidAdjustmentReason(String reason) =>
      adjustmentReasons.contains(reason.toLowerCase());

  static Result<void> validateAdjustment({
    required String reason,
    required int deltaQty,
    String? note,
  }) {
    if (deltaQty == 0) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Adjustment qty cannot be zero',
        ),
      );
    }
    if (!isValidAdjustmentReason(reason)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid adjustment reason',
          details: {'reason': reason},
        ),
      );
    }
    if (reason.toLowerCase() == 'other' &&
        (note == null || note.trim().isEmpty)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Note required for other adjustments',
        ),
      );
    }
    return const Success(null);
  }

  static bool isLowStock({required int onHand, required int reorderPoint}) =>
      onHand > 0 && onHand <= reorderPoint;

  static bool isOutOfStock(int onHand) => onHand <= 0;

  static Result<void> validateAdjust({
    required int delta,
    required String reasonCode,
    String? notes,
    required int currentOnHand,
  }) {
    final base = validateAdjustment(
      reason: reasonCode,
      deltaQty: delta,
      note: notes,
    );
    if (base.isFailure) return base;
    if (currentOnHand + delta < 0) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Adjustment would make stock negative',
          details: {'on_hand': currentOnHand, 'delta': delta},
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validateCycleCountQty({required int countedQty}) {
    if (countedQty < 0) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Counted quantity cannot be negative',
        ),
      );
    }
    return const Success(null);
  }

  static bool preferFefo({
    required DateTime? candidateExpiry,
    required DateTime? otherExpiry,
  }) {
    if (candidateExpiry == null) return false;
    if (otherExpiry == null) return true;
    return candidateExpiry.isBefore(otherExpiry);
  }

  /// Prefer earliest expiry (FEFO). Null expiry sorts last.
  static int compareFefo(DateTime? a, DateTime? b) {
    if (a == null && b == null) return 0;
    if (a == null) return 1;
    if (b == null) return -1;
    return a.compareTo(b);
  }

  /// Returns lot ids ordered by FEFO from [lots] maps with `lot_id` + `expires_at`.
  static List<String> preferFefoLotIds(List<Map<String, dynamic>> lots) {
    final sorted = [...lots]..sort((a, b) {
        final aExp = DateTime.tryParse(a['expires_at']?.toString() ?? '');
        final bExp = DateTime.tryParse(b['expires_at']?.toString() ?? '');
        return compareFefo(aExp, bExp);
      });
    return sorted
        .map((e) => e['lot_id']?.toString() ?? e['id']?.toString() ?? '')
        .where((id) => id.isNotEmpty)
        .toList();
  }
}
