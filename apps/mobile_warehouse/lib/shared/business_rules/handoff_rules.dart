import 'package:nexora_core/nexora_core.dart';

import '../../features/dispatch/domain/entities/handoff_task.dart';

/// Courier handoff QR token matching rules.
abstract final class HandoffRules {
  static const failReasonCodes = <String>{
    'qr_mismatch',
    'courier_timeout',
    'order_mismatch',
    'damaged',
    'other',
  };

  static const Map<HandoffStatus, Set<HandoffStatus>> _allowed = {
    HandoffStatus.waitingCourier: {
      HandoffStatus.courierArrived,
      HandoffStatus.failedPickup,
    },
    HandoffStatus.courierArrived: {
      HandoffStatus.verifying,
      HandoffStatus.failedPickup,
    },
    HandoffStatus.verifying: {
      HandoffStatus.handedOff,
      HandoffStatus.failedPickup,
    },
    HandoffStatus.handedOff: {},
    HandoffStatus.failedPickup: {},
  };

  static bool canMarkArrived(HandoffStatus status) =>
      status == HandoffStatus.waitingCourier;
  static bool canScanQr(HandoffStatus status) =>
      status == HandoffStatus.courierArrived ||
      status == HandoffStatus.verifying;
  static bool canConfirm(HandoffStatus status) =>
      status == HandoffStatus.verifying;
  static bool canFail(HandoffStatus status) =>
      status != HandoffStatus.handedOff && status != HandoffStatus.failedPickup;

  static bool canTransition(HandoffStatus from, HandoffStatus to) {
    if (from == to) return true;
    return _allowed[from]?.contains(to) ?? false;
  }

  static Result<void> validateTransition({
    required HandoffStatus from,
    required HandoffStatus to,
  }) {
    if (canTransition(from, to)) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Illegal handoff workflow transition',
        details: {'from': from.name, 'to': to.name},
      ),
    );
  }

  static Result<void> validateHandoffScan({
    required HandoffStatus status,
    required String scannedToken,
    required String expectedToken,
    required String expectedOrderId,
    String? scannedOrderId,
    String? expectedCourierId,
  }) {
    if (!canScanQr(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Handoff scan is not allowed in current status',
          details: {'status': status.name},
        ),
      );
    }
    if (scannedToken.trim().isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Scanned handoff token is empty',
        ),
      );
    }
    // Prefer structured token parse when courier id known; else exact match.
    if (expectedCourierId != null && expectedCourierId.isNotEmpty) {
      return validateTokenMatch(
        scannedToken: scannedToken,
        expectedOrderId: expectedOrderId,
        expectedCourierId: expectedCourierId,
      );
    }
    if (scannedToken.trim() != expectedToken.trim()) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Handoff QR does not match',
        ),
      );
    }
    if (scannedOrderId != null &&
        scannedOrderId.trim().isNotEmpty &&
        scannedOrderId.trim() != expectedOrderId.trim()) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Order id on QR does not match handoff',
          details: {
            'expected_order_id': expectedOrderId,
            'scanned_order_id': scannedOrderId,
          },
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validateFailReason(String reasonCode, {String? notes}) {
    if (!failReasonCodes.contains(reasonCode)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid handoff fail reason',
          details: {'reason_code': reasonCode},
        ),
      );
    }
    if (reasonCode == 'other' && (notes == null || notes.trim().isEmpty)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Notes are required for fail reason "other"',
        ),
      );
    }
    return const Success(null);
  }

  /// Expected payload shape: orderId + courierId (+ optional handoffId).
  static bool tokenMatches({
    required String scannedToken,
    required String expectedOrderId,
    required String expectedCourierId,
    String? expectedHandoffId,
  }) {
    final parsed = parseHandoffToken(scannedToken);
    if (parsed == null) return false;
    if (parsed.orderId != expectedOrderId) return false;
    if (parsed.courierId != expectedCourierId) return false;
    if (expectedHandoffId != null &&
        expectedHandoffId.isNotEmpty &&
        parsed.handoffId != null &&
        parsed.handoffId != expectedHandoffId) {
      return false;
    }
    return true;
  }

  static Result<void> validateTokenMatch({
    required String scannedToken,
    required String expectedOrderId,
    required String expectedCourierId,
    String? expectedHandoffId,
  }) {
    if (tokenMatches(
      scannedToken: scannedToken,
      expectedOrderId: expectedOrderId,
      expectedCourierId: expectedCourierId,
      expectedHandoffId: expectedHandoffId,
    )) {
      return const Success(null);
    }
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Handoff QR does not match courier and order',
        details: {
          'expected_order_id': expectedOrderId,
          'expected_courier_id': expectedCourierId,
          if (expectedHandoffId != null) 'expected_handoff_id': expectedHandoffId,
        },
      ),
    );
  }

  /// Supports `orderId|courierId` or `orderId|courierId|handoffId`
  /// or JSON-ish `order_id=..&courier_id=..`.
  static HandoffTokenParts? parseHandoffToken(String raw) {
    final trimmed = raw.trim();
    if (trimmed.isEmpty) return null;

    if (trimmed.contains('|')) {
      final parts = trimmed.split('|');
      if (parts.length < 2) return null;
      return HandoffTokenParts(
        orderId: parts[0].trim(),
        courierId: parts[1].trim(),
        handoffId: parts.length > 2 ? parts[2].trim() : null,
      );
    }

    final query = Uri.tryParse(
      trimmed.contains('://') ? trimmed : 'nexora://handoff?$trimmed',
    );
    if (query != null && query.queryParameters.isNotEmpty) {
      final orderId =
          query.queryParameters['order_id'] ?? query.queryParameters['orderId'];
      final courierId = query.queryParameters['courier_id'] ??
          query.queryParameters['courierId'];
      if (orderId == null || courierId == null) return null;
      return HandoffTokenParts(
        orderId: orderId,
        courierId: courierId,
        handoffId: query.queryParameters['handoff_id'] ??
            query.queryParameters['handoffId'],
      );
    }

    return null;
  }
}

class HandoffTokenParts {
  const HandoffTokenParts({
    required this.orderId,
    required this.courierId,
    this.handoffId,
  });

  final String orderId;
  final String courierId;
  final String? handoffId;
}
