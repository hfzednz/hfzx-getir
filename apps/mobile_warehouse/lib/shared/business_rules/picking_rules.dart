import 'package:nexora_core/nexora_core.dart';

import '../../features/picking/domain/entities/pick_task.dart';

/// Picking claim / scan / short-pick rules (ARCHITECTURE.md).
abstract final class PickingRules {
  static const shortPickReasons = <String>{
    'missing_stock',
    'damaged',
    'wrong_bin',
    'expired',
    'inaccessible',
    'other',
  };

  // --- Status gates (used by picking UI) ---

  static bool canClaim(PickTaskStatus status) =>
      status == PickTaskStatus.queued;

  static bool canStart(PickTaskStatus status) =>
      status == PickTaskStatus.claimed;

  static bool canScanLine(PickTaskStatus status) =>
      status == PickTaskStatus.inProgress ||
      status == PickTaskStatus.shortPick;

  static bool canShortPick(PickTaskStatus status) =>
      status == PickTaskStatus.inProgress;

  static bool canComplete(PickTask task) {
    if (task.status != PickTaskStatus.inProgress &&
        task.status != PickTaskStatus.shortPick) {
      return false;
    }
    return task.lines.every((l) => l.isComplete);
  }

  static bool canStage(PickTaskStatus status) =>
      status == PickTaskStatus.picked;

  static const Map<PickTaskStatus, Set<PickTaskStatus>> _allowed = {
    PickTaskStatus.queued: {PickTaskStatus.claimed, PickTaskStatus.exception},
    PickTaskStatus.claimed: {
      PickTaskStatus.inProgress,
      PickTaskStatus.queued,
      PickTaskStatus.exception,
    },
    PickTaskStatus.inProgress: {
      PickTaskStatus.picked,
      PickTaskStatus.shortPick,
      PickTaskStatus.exception,
    },
    PickTaskStatus.shortPick: {
      PickTaskStatus.inProgress,
      PickTaskStatus.exception,
      PickTaskStatus.picked,
    },
    PickTaskStatus.picked: {PickTaskStatus.staged},
    PickTaskStatus.staged: {},
    PickTaskStatus.exception: {PickTaskStatus.queued},
  };

  static bool canTransition(PickTaskStatus from, PickTaskStatus to) {
    if (from == to) return true;
    return _allowed[from]?.contains(to) ?? false;
  }

  static Result<void> validateTransition({
    required PickTaskStatus from,
    required PickTaskStatus to,
  }) {
    if (canTransition(from, to)) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Illegal pick workflow transition',
        details: {'from': from.name, 'to': to.name},
      ),
    );
  }

  /// Claim is allowed only when task is queued / unassigned (string API).
  static bool canClaimRaw({
    required String status,
    String? assignedUserId,
    String? currentUserId,
  }) {
    final normalized = status.toLowerCase();
    if (normalized != 'queued' && normalized != 'pending') return false;
    if (assignedUserId == null || assignedUserId.isEmpty) return true;
    return assignedUserId == currentUserId;
  }

  static Result<void> validateClaim({
    required String status,
    String? assignedUserId,
    String? currentUserId,
  }) {
    if (canClaimRaw(
      status: status,
      assignedUserId: assignedUserId,
      currentUserId: currentUserId,
    )) {
      return const Success(null);
    }
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Cannot claim pick task',
        details: {
          'status': status,
          'assigned_user_id': assignedUserId,
        },
      ),
    );
  }

  /// Scanned SKU must match the expected line SKU and bin.
  static bool scanMatchesLine({
    required String scannedSku,
    required String expectedSku,
    required String scannedBinId,
    required String expectedBinId,
  }) {
    return scannedSku.trim().toUpperCase() == expectedSku.trim().toUpperCase() &&
        scannedBinId.trim().toUpperCase() == expectedBinId.trim().toUpperCase();
  }

  static Result<void> validateScan({
    required String scannedSku,
    required String expectedSku,
    required String scannedBinId,
    required String expectedBinId,
  }) {
    if (scanMatchesLine(
      scannedSku: scannedSku,
      expectedSku: expectedSku,
      scannedBinId: scannedBinId,
      expectedBinId: expectedBinId,
    )) {
      return const Success(null);
    }
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Scan must match line SKU and bin',
        details: {
          'scanned_sku': scannedSku,
          'expected_sku': expectedSku,
          'scanned_bin_id': scannedBinId,
          'expected_bin_id': expectedBinId,
        },
      ),
    );
  }

  /// Line scan: barcode must match line barcode/SKU while task is in progress.
  static Result<void> validateLineScan({
    required PickTaskStatus status,
    required PickLine line,
    required String scannedBarcode,
    int qtyDelta = 1,
  }) {
    if (!canScanLine(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Cannot scan line in status ${status.wireName}',
        ),
      );
    }
    if (line.isComplete) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Line already complete',
        ),
      );
    }
    final scanned = scannedBarcode.trim().toUpperCase();
    final matches = scanned == line.barcode.trim().toUpperCase() ||
        scanned == line.sku.trim().toUpperCase() ||
        (line.substitutionSku != null &&
            scanned == line.substitutionSku!.trim().toUpperCase());
    if (!matches) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Scan must match line SKU/barcode',
          details: {
            'scanned': scannedBarcode,
            'expected_sku': line.sku,
            'expected_barcode': line.barcode,
            'bin': line.bin,
          },
        ),
      );
    }
    if (qtyDelta <= 0) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Scan qty must be positive',
        ),
      );
    }
    if (line.pickedQty + qtyDelta > line.qty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Scan qty exceeds remaining line qty',
        ),
      );
    }
    return const Success(null);
  }

  static bool isValidShortPickReason(String reason) =>
      shortPickReasons.contains(reason.toLowerCase());

  static Result<void> validateShortPick({
    PickTaskStatus? status,
    PickLine? line,
    int? missingQty,
    String? reason,
    int? expectedQty,
    int? pickedQty,
  }) {
    if (status != null && !canShortPick(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Cannot short-pick in status ${status.wireName}',
        ),
      );
    }

    final expected = expectedQty ?? line?.qty;
    final picked = pickedQty ?? line?.pickedQty;
    final missing = missingQty ??
        (expected != null && picked != null ? expected - picked : null);

    if (expected != null && picked != null && picked >= expected) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Short-pick requires picked qty less than expected',
        ),
      );
    }
    if (missing != null && missing <= 0) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Short-pick missing qty must be positive',
        ),
      );
    }
    if (reason != null && !isValidShortPickReason(reason)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid short-pick reason',
          details: {'reason': reason},
        ),
      );
    }
    return const Success(null);
  }
}
