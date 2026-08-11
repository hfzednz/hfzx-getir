import 'dart:async';

import 'package:battery_plus/battery_plus.dart';
import 'package:geolocator/geolocator.dart';
import 'package:uuid/uuid.dart';

import '../../data/local/courier_local_store.dart';
import '../business_rules/location_rules.dart';

enum LocationTrackingMode { idle, activeDelivery, lowBattery }

/// Geolocator wrapper with adaptive intervals + breadcrumb enqueue.
class LocationService {
  LocationService({
    required CourierLocalStore localStore,
    Battery? battery,
  })  : _localStore = localStore,
        _battery = battery ?? Battery();

  final CourierLocalStore _localStore;
  final Battery _battery;
  final _uuid = const Uuid();

  StreamSubscription<Position>? _subscription;
  Position? _lastPosition;
  LocationTrackingMode _mode = LocationTrackingMode.idle;
  bool _activeDelivery = false;

  Position? get lastPosition => _lastPosition;

  Future<bool> ensurePermission() async {
    var permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
    }
    if (permission == LocationPermission.deniedForever ||
        permission == LocationPermission.denied) {
      return false;
    }
    return true;
  }

  Future<void> start({required bool activeDelivery}) async {
    _activeDelivery = activeDelivery;
    final ok = await ensurePermission();
    if (!ok) return;

    await stop();
    final settings = await _settingsForCurrentContext();
    _subscription = Geolocator.getPositionStream(locationSettings: settings)
        .listen(_onPosition);
  }

  Future<void> setActiveDelivery(bool value) async {
    if (_activeDelivery == value) return;
    _activeDelivery = value;
    if (_subscription != null) {
      await start(activeDelivery: value);
    }
  }

  Future<void> stop() async {
    await _subscription?.cancel();
    _subscription = null;
  }

  Future<LocationSettings> _settingsForCurrentContext() async {
    final level = await _battery.batteryLevel;
    final lowBattery = level <= 15;
    _mode = lowBattery
        ? LocationTrackingMode.lowBattery
        : (_activeDelivery
            ? LocationTrackingMode.activeDelivery
            : LocationTrackingMode.idle);

    final interval = switch (_mode) {
      LocationTrackingMode.activeDelivery => const Duration(seconds: 5),
      LocationTrackingMode.lowBattery => const Duration(seconds: 90),
      LocationTrackingMode.idle => const Duration(seconds: 45),
    };
    // Adaptive interval informs distance filter aggressiveness.
    final distanceFilter = switch (_mode) {
      LocationTrackingMode.activeDelivery => 8,
      LocationTrackingMode.lowBattery => 50,
      LocationTrackingMode.idle => interval.inSeconds >= 45 ? 25 : 15,
    };

    return LocationSettings(
      accuracy: _mode == LocationTrackingMode.activeDelivery
          ? LocationAccuracy.high
          : LocationAccuracy.medium,
      distanceFilter: distanceFilter,
    );
  }

  Future<void> _onPosition(Position position) async {
    final previous = _lastPosition;
    if (previous != null) {
      final jump = LocationRules.detectImpossibleJump(
        previousLat: previous.latitude,
        previousLng: previous.longitude,
        previousAt: previous.timestamp,
        nextLat: position.latitude,
        nextLng: position.longitude,
        nextAt: position.timestamp,
      );
      if (jump.isSuspicious) {
        // Still record; flag for integrity review.
        await _localStore.enqueueBreadcrumb({
          'id': _uuid.v4(),
          'lat': position.latitude,
          'lng': position.longitude,
          'accuracy_m': position.accuracy,
          'speed_mps': position.speed,
          'recorded_at': position.timestamp.toUtc().toIso8601String(),
          'spoof_flag': true,
          'max_plausible_mps': jump.maxPlausibleSpeedMps,
          'observed_mps': jump.observedSpeedMps,
        });
        _lastPosition = position;
        return;
      }
    }

    _lastPosition = position;
    await _localStore.enqueueBreadcrumb({
      'id': _uuid.v4(),
      'lat': position.latitude,
      'lng': position.longitude,
      'accuracy_m': position.accuracy,
      'speed_mps': position.speed,
      'recorded_at': position.timestamp.toUtc().toIso8601String(),
      'spoof_flag': false,
      'mode': _mode.name,
    });
  }

  Future<Position?> currentPosition() async {
    final ok = await ensurePermission();
    if (!ok) return null;
    return Geolocator.getCurrentPosition(
      locationSettings: const LocationSettings(accuracy: LocationAccuracy.high),
    );
  }
}
