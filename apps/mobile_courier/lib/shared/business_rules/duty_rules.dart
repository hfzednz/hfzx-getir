import 'package:nexora_core/nexora_core.dart';

import '../../features/status/domain/entities/duty_status.dart';

/// Legal duty status transitions (ARCHITECTURE.md).
///
/// ```text
/// offline → online → (busy | on_break | emergency)
/// ```
abstract final class DutyRules {
  static const Map<DutyStatus, Set<DutyStatus>> _allowed = {
    DutyStatus.offline: {DutyStatus.online, DutyStatus.emergency},
    DutyStatus.online: {
      DutyStatus.offline,
      DutyStatus.busy,
      DutyStatus.onBreak,
      DutyStatus.emergency,
    },
    DutyStatus.busy: {
      DutyStatus.online,
      DutyStatus.onBreak,
      DutyStatus.emergency,
      DutyStatus.offline,
    },
    DutyStatus.onBreak: {
      DutyStatus.online,
      DutyStatus.busy,
      DutyStatus.emergency,
      DutyStatus.offline,
    },
    DutyStatus.emergency: {
      DutyStatus.offline,
      DutyStatus.online,
    },
  };

  static bool canTransition(
    DutyStatus from,
    DutyStatus to, {
    bool hasActiveDelivery = false,
  }) {
    if (from == to) return true;
    if (hasActiveDelivery && to == DutyStatus.offline) {
      return false;
    }
    if (hasActiveDelivery && to == DutyStatus.onBreak) {
      return false;
    }
    return _allowed[from]?.contains(to) ?? false;
  }

  static Result<void> validateTransition({
    required DutyStatus from,
    required DutyStatus to,
    bool hasActiveDelivery = false,
  }) {
    if (canTransition(from, to, hasActiveDelivery: hasActiveDelivery)) {
      return const Success(null);
    }
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Illegal duty status transition',
        details: {
          'from': from.name,
          'to': to.name,
          'has_active_delivery': hasActiveDelivery,
        },
      ),
    );
  }

  /// Couriers must be online (or busy with an active job) to accept offers.
  static bool canAcceptOffers(DutyStatus status) =>
      status == DutyStatus.online || status == DutyStatus.busy;
}
