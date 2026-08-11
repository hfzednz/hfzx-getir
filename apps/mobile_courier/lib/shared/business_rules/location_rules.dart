import 'dart:math' as math;

import 'package:nexora_core/nexora_core.dart';

class LocationJumpResult {
  const LocationJumpResult({
    required this.isSuspicious,
    required this.observedSpeedMps,
    required this.maxPlausibleSpeedMps,
    required this.distanceMeters,
  });

  final bool isSuspicious;
  final double observedSpeedMps;
  final double maxPlausibleSpeedMps;
  final double distanceMeters;
}

/// Spoof / impossible jump detection + battery adaptive helpers.
abstract final class LocationRules {
  /// Max plausible ground speed for courier vehicles (m/s). ~120 km/h.
  static const maxPlausibleSpeedMps = 33.3;

  /// Ignore tiny GPS jitter under this distance (meters).
  static const minDistanceMeters = 25.0;

  /// Minimum elapsed time to evaluate speed (seconds).
  static const minElapsedSeconds = 2.0;

  /// Battery % at/below which location updates slow down.
  static const lowBatteryThresholdPercent = 15;

  /// Idle online ping interval.
  static const Duration idleInterval = Duration(seconds: 45);

  /// Active delivery ping interval.
  static const Duration activeInterval = Duration(seconds: 5);

  /// Low-battery ping interval.
  static const Duration lowBatteryInterval = Duration(seconds: 90);

  static bool isLowBattery(int batteryPercent) =>
      batteryPercent <= lowBatteryThresholdPercent;

  static Duration pingInterval({
    required bool hasActiveDelivery,
    required int batteryLevel,
  }) {
    if (isLowBattery(batteryLevel)) return lowBatteryInterval;
    if (hasActiveDelivery) return activeInterval;
    return idleInterval;
  }

  static double distanceMeters({
    required double lat1,
    required double lng1,
    required double lat2,
    required double lng2,
  }) =>
      haversineMeters(lat1: lat1, lng1: lng1, lat2: lat2, lng2: lng2);

  static double haversineMeters({
    required double lat1,
    required double lng1,
    required double lat2,
    required double lng2,
  }) {
    const earthRadius = 6371000.0;
    final dLat = _toRad(lat2 - lat1);
    final dLng = _toRad(lng2 - lng1);
    final a = math.sin(dLat / 2) * math.sin(dLat / 2) +
        math.cos(_toRad(lat1)) *
            math.cos(_toRad(lat2)) *
            math.sin(dLng / 2) *
            math.sin(dLng / 2);
    final c = 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a));
    return earthRadius * c;
  }

  static LocationJumpResult detectImpossibleJump({
    required double previousLat,
    required double previousLng,
    required DateTime previousAt,
    required double nextLat,
    required double nextLng,
    required DateTime nextAt,
    double maxSpeedMps = maxPlausibleSpeedMps,
  }) {
    final distance = haversineMeters(
      lat1: previousLat,
      lng1: previousLng,
      lat2: nextLat,
      lng2: nextLng,
    );
    final elapsed = nextAt.difference(previousAt).inMilliseconds / 1000.0;

    if (distance < minDistanceMeters || elapsed < minElapsedSeconds) {
      return LocationJumpResult(
        isSuspicious: false,
        observedSpeedMps: 0,
        maxPlausibleSpeedMps: maxSpeedMps,
        distanceMeters: distance,
      );
    }

    final observed = distance / elapsed;
    return LocationJumpResult(
      isSuspicious: observed > maxSpeedMps,
      observedSpeedMps: observed,
      maxPlausibleSpeedMps: maxSpeedMps,
      distanceMeters: distance,
    );
  }

  static Result<void> validatePing({
    required double lat,
    required double lng,
    DateTime? at,
    double? previousLat,
    double? previousLng,
    DateTime? previousAt,
    double? nextLat,
    double? nextLng,
    DateTime? nextAt,
  }) {
    if (lat < -90 || lat > 90 || lng < -180 || lng > 180) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Invalid coordinates',
          details: {'lat': lat, 'lng': lng},
        ),
      );
    }

    final fromLat = previousLat;
    final fromLng = previousLng;
    final fromAt = previousAt;
    final toLat = nextLat ?? lat;
    final toLng = nextLng ?? lng;
    final toAt = nextAt ?? at;

    if (fromLat == null || fromLng == null || fromAt == null || toAt == null) {
      return const Success(null);
    }

    final jump = detectImpossibleJump(
      previousLat: fromLat,
      previousLng: fromLng,
      previousAt: fromAt,
      nextLat: toLat,
      nextLng: toLng,
      nextAt: toAt,
    );
    if (!jump.isSuspicious) return const Success(null);
    return Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Impossible location jump detected',
        details: {
          'distance_m': jump.distanceMeters,
          'observed_mps': jump.observedSpeedMps,
          'max_plausible_mps': jump.maxPlausibleSpeedMps,
        },
      ),
    );
  }

  static double _toRad(double deg) => deg * math.pi / 180.0;
}
