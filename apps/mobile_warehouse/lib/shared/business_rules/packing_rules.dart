import 'package:nexora_core/nexora_core.dart';

import '../../features/packing/domain/entities/pack_task.dart';

/// Packing weight tolerance and label requirements.
abstract final class PackingRules {
  /// Default relative weight tolerance (±%).
  static const defaultWeightTolerancePercent = 5.0;

  /// Absolute grams floor when expected weight is very small.
  static const minAbsoluteToleranceGrams = 10.0;

  static bool isWeightWithinTolerance({
    required double actualGrams,
    required double expectedGrams,
    double tolerancePercent = defaultWeightTolerancePercent,
  }) {
    if (expectedGrams <= 0) return actualGrams >= 0;
    final relative = expectedGrams * (tolerancePercent / 100);
    final allowed = relative < minAbsoluteToleranceGrams
        ? minAbsoluteToleranceGrams
        : relative;
    return (actualGrams - expectedGrams).abs() <= allowed;
  }

  static Result<void> validateWeight({
    required double actualGrams,
    required double expectedGrams,
    double tolerancePercent = defaultWeightTolerancePercent,
  }) {
    if (actualGrams <= 0) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Weight must be positive',
        ),
      );
    }
    if (isWeightWithinTolerance(
      actualGrams: actualGrams,
      expectedGrams: expectedGrams,
      tolerancePercent: tolerancePercent,
    )) {
      return const Success(null);
    }
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Weight outside tolerance',
        details: {
          'actual_grams': actualGrams,
          'expected_grams': expectedGrams,
          'tolerance_percent': tolerancePercent,
        },
      ),
    );
  }

  /// Pack seal requires a printed shipping / bag label.
  static bool hasRequiredLabel({
    required bool labelPrinted,
    String? labelId,
  }) {
    if (!labelPrinted) return false;
    return labelId != null && labelId.trim().isNotEmpty;
  }

  static Result<void> validateRequiredLabel({
    required bool labelPrinted,
    String? labelId,
  }) {
    if (hasRequiredLabel(labelPrinted: labelPrinted, labelId: labelId)) {
      return const Success(null);
    }
    return const Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Pack requires a printed label before seal',
      ),
    );
  }

  static const Map<PackTaskStatus, Set<PackTaskStatus>> _allowed = {
    PackTaskStatus.readyToPack: {PackTaskStatus.packing, PackTaskStatus.qcHold},
    PackTaskStatus.packing: {PackTaskStatus.weighed, PackTaskStatus.qcHold},
    PackTaskStatus.weighed: {PackTaskStatus.labeled, PackTaskStatus.qcHold},
    PackTaskStatus.labeled: {PackTaskStatus.packed, PackTaskStatus.qcHold},
    PackTaskStatus.packed: {PackTaskStatus.dispatchQueued},
    PackTaskStatus.qcHold: {PackTaskStatus.readyToPack, PackTaskStatus.packing},
    PackTaskStatus.dispatchQueued: {},
  };

  static bool canClaim(PackTaskStatus status) =>
      status == PackTaskStatus.readyToPack;
  static bool canWeigh(PackTaskStatus status) =>
      status == PackTaskStatus.packing;
  static bool canPrintLabel(PackTaskStatus status) =>
      status == PackTaskStatus.weighed || status == PackTaskStatus.labeled;
  static bool canSeal(PackTaskStatus status) =>
      status == PackTaskStatus.labeled;
  static bool canComplete(PackTaskStatus status) =>
      status == PackTaskStatus.packed;

  static bool canTransition(PackTaskStatus from, PackTaskStatus to) {
    if (from == to) return true;
    return _allowed[from]?.contains(to) ?? false;
  }

  static Result<void> validateTransition({
    required PackTaskStatus from,
    required PackTaskStatus to,
  }) {
    if (canTransition(from, to)) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Illegal pack workflow transition',
        details: {'from': from.name, 'to': to.name},
      ),
    );
  }

  static Result<void> validateSeal({
    required PackTaskStatus status,
    required bool sealed,
    required bool labelPrinted,
    String? labelId,
  }) {
    if (!canSeal(status) && status != PackTaskStatus.labeled) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Seal is not allowed in current status',
          details: {'status': status.name},
        ),
      );
    }
    final labelCheck = validateRequiredLabel(
      labelPrinted: labelPrinted,
      labelId: labelId ?? (labelPrinted ? 'printed' : null),
    );
    if (labelCheck.isFailure) return labelCheck;
    if (!sealed) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Package must be sealed',
        ),
      );
    }
    return const Success(null);
  }
}
