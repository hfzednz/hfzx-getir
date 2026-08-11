import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_courier/shared/business_rules/location_rules.dart';

void main() {
  group('LocationRules.pingInterval', () {
    test('uses active interval during delivery', () {
      expect(
        LocationRules.pingInterval(hasActiveDelivery: true, batteryLevel: 80),
        LocationRules.activeInterval,
      );
    });

    test('slows on low battery', () {
      expect(
        LocationRules.pingInterval(hasActiveDelivery: true, batteryLevel: 10),
        LocationRules.lowBatteryInterval,
      );
    });
  });

  group('LocationRules.validatePing', () {
    test('rejects invalid coordinates', () {
      final result = LocationRules.validatePing(lat: 200, lng: 0);
      expect(result.isFailure, isTrue);
    });

    test('allows first ping without previous', () {
      final result = LocationRules.validatePing(lat: 41.0, lng: 29.0);
      expect(result.isSuccess, isTrue);
    });

    test('detects impossible jump', () {
      final previousAt = DateTime.utc(2026, 1, 1, 12);
      final result = LocationRules.validatePing(
        lat: 41.0,
        lng: 29.0,
        previousLat: 40.0,
        previousLng: 29.0,
        previousAt: previousAt,
        at: previousAt.add(const Duration(seconds: 5)),
      );
      expect(result.isFailure, isTrue);
    });

    test('allows plausible movement', () {
      final previousAt = DateTime.utc(2026, 1, 1, 12);
      final result = LocationRules.validatePing(
        lat: 41.001,
        lng: 29.0,
        previousLat: 41.0,
        previousLng: 29.0,
        previousAt: previousAt,
        at: previousAt.add(const Duration(seconds: 30)),
      );
      expect(result.isSuccess, isTrue);
    });
  });

  group('LocationRules.distanceMeters', () {
    test('returns ~0 for same point', () {
      expect(
        LocationRules.distanceMeters(
          lat1: 41,
          lng1: 29,
          lat2: 41,
          lng2: 29,
        ),
        closeTo(0, 0.01),
      );
    });
  });
}
