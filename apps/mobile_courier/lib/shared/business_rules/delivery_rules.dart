import 'package:nexora_core/nexora_core.dart';

import '../../features/deliveries/domain/entities/delivery_job.dart';

/// Accept window, POD required fields, fail reasons, workflow transitions.
abstract final class DeliveryRules {
  /// Default offer accept TTL (seconds) when server does not specify.
  static const defaultAcceptWindowSeconds = 30;

  static const failureReasonCodes = <String>{
    'customer_unavailable',
    'wrong_address',
    'refused',
    'unsafe_location',
    'damaged_goods',
    'other',
  };

  static const Map<DeliveryLifecycleStatus, Set<DeliveryLifecycleStatus>>
      _workflowAllowed = {
    DeliveryLifecycleStatus.assigned: {
      DeliveryLifecycleStatus.enRouteStore,
      DeliveryLifecycleStatus.failed,
      DeliveryLifecycleStatus.cancelled,
    },
    DeliveryLifecycleStatus.enRouteStore: {
      DeliveryLifecycleStatus.atStore,
      DeliveryLifecycleStatus.failed,
    },
    DeliveryLifecycleStatus.atStore: {
      DeliveryLifecycleStatus.pickedUp,
      DeliveryLifecycleStatus.failed,
    },
    DeliveryLifecycleStatus.pickedUp: {
      DeliveryLifecycleStatus.enRouteCustomer,
      DeliveryLifecycleStatus.failed,
    },
    DeliveryLifecycleStatus.enRouteCustomer: {
      DeliveryLifecycleStatus.arrived,
      DeliveryLifecycleStatus.failed,
    },
    DeliveryLifecycleStatus.arrived: {
      DeliveryLifecycleStatus.delivered,
      DeliveryLifecycleStatus.failed,
    },
    DeliveryLifecycleStatus.delivered: {},
    DeliveryLifecycleStatus.failed: {},
    DeliveryLifecycleStatus.cancelled: {},
  };

  static const requiredPodFields = {
    'photo_uri',
    'recipient_name',
    'delivered_at',
  };

  static bool isTerminal(DeliveryLifecycleStatus status) =>
      status == DeliveryLifecycleStatus.delivered ||
      status == DeliveryLifecycleStatus.failed ||
      status == DeliveryLifecycleStatus.cancelled;

  static bool canScanPickup(DeliveryLifecycleStatus status) =>
      status == DeliveryLifecycleStatus.atStore;

  static bool canSubmitPod(DeliveryLifecycleStatus status) =>
      status == DeliveryLifecycleStatus.arrived;

  static bool canMarkFailed(DeliveryLifecycleStatus status) =>
      !isTerminal(status);

  static bool canTransition(
    DeliveryLifecycleStatus from,
    DeliveryLifecycleStatus to,
  ) {
    if (from == to) return true;
    return _workflowAllowed[from]?.contains(to) ?? false;
  }

  static Result<void> validateTransition({
    required DeliveryLifecycleStatus from,
    required DeliveryLifecycleStatus to,
  }) {
    if (canTransition(from, to)) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Illegal delivery workflow transition',
        details: {'from': from.name, 'to': to.name},
      ),
    );
  }

  static Result<void> validateAcceptWindow({
    required DateTime offeredAt,
    required DateTime now,
    int acceptWindowSeconds = defaultAcceptWindowSeconds,
    DateTime? expiresAt,
  }) {
    final deadline =
        expiresAt ?? offeredAt.add(Duration(seconds: acceptWindowSeconds));
    if (now.isAfter(deadline)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Offer accept window expired',
          details: {
            'offered_at': offeredAt.toIso8601String(),
            'expires_at': deadline.toIso8601String(),
          },
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validatePickupScan({
    required DeliveryLifecycleStatus status,
    required String scannedToken,
    required String expectedHandoffToken,
  }) {
    if (!canScanPickup(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Pickup scan is not allowed in current status',
          details: {'status': status.name},
        ),
      );
    }
    if (scannedToken.trim().isEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Scanned token is empty',
          details: {'field': 'scanned_token'},
        ),
      );
    }
    if (scannedToken.trim() != expectedHandoffToken.trim()) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Handoff QR does not match this delivery',
          details: {'field': 'handoff_token'},
        ),
      );
    }
    return const Success(null);
  }

  /// POD validation — supports both payload map and convenience flags.
  static Result<void> validatePod({
    DeliveryLifecycleStatus? status,
    Map<String, dynamic>? payload,
    bool? hasPhoto,
    String? otp,
    bool otpRequired = false,
    bool signatureRequired = false,
  }) {
    if (status != null && !canSubmitPod(status)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'POD is not allowed in current status',
          details: {'status': status.name},
        ),
      );
    }

    if (payload != null) {
      final missing = <String>[];
      for (final field in requiredPodFields) {
        final value = payload[field];
        if (value == null || (value is String && value.trim().isEmpty)) {
          missing.add(field);
        }
      }
      if (signatureRequired) {
        final sig = payload['signature_uri'];
        if (sig == null || (sig is String && sig.trim().isEmpty)) {
          missing.add('signature_uri');
        }
      }
      if (otpRequired) {
        final code = payload['otp_code'] ?? payload['otp'];
        if (code == null || (code is String && code.trim().isEmpty)) {
          missing.add('otp_code');
        }
      }
      if (missing.isNotEmpty) {
        return Failure(
          NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Proof of delivery is incomplete',
            details: {'missing_fields': missing},
          ),
        );
      }
      return const Success(null);
    }

    if (hasPhoto == false) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery photo is required',
          details: {'field': 'photo'},
        ),
      );
    }
    if (otpRequired && (otp == null || otp.trim().isEmpty)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery OTP is required',
          details: {'field': 'otp'},
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validateFailureReason(String reasonCode) {
    if (!failureReasonCodes.contains(reasonCode)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid failure reason',
          details: {'reason_code': reasonCode},
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validateFail({
    required String reasonCode,
    String? notes,
  }) {
    final base = validateFailureReason(reasonCode);
    if (base.isFailure) return base;
    if (reasonCode == 'other' && (notes == null || notes.trim().isEmpty)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Notes are required for fail reason "other"',
          details: {'field': 'notes'},
        ),
      );
    }
    return const Success(null);
  }
}
