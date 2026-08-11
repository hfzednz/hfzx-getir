import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/location_rules.dart';
import '../../data/datasources/navigation_remote_datasource.dart';
import '../../data/repositories/navigation_repository_impl.dart';
import '../../domain/entities/route_entity.dart';
import '../../domain/repositories/navigation_repository.dart';

final navigationRemoteDataSourceProvider =
    Provider<NavigationRemoteDataSource>((ref) {
  return NavigationRemoteDataSource(ref.watch(apiClientProvider));
});

final navigationRepositoryProvider = Provider<NavigationRepository>((ref) {
  return NavigationRepositoryImpl(ref.watch(navigationRemoteDataSourceProvider));
});

final deliveryRouteProvider =
    FutureProvider.autoDispose.family<DeliveryRoute, String>((ref, id) async {
  final result = await ref.watch(navigationRepositoryProvider).fetchRoute(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final locationPingProvider = Provider((ref) => LocationPingActions(ref));

class LocationPingActions {
  LocationPingActions(this._ref);
  final Ref _ref;

  double? _prevLat;
  double? _prevLng;
  DateTime? _prevAt;

  Future<bool> ping({
    required double lat,
    required double lng,
    double? accuracyMeters,
  }) async {
    final now = DateTime.now();
    final validation = LocationRules.validatePing(
      lat: lat,
      lng: lng,
      previousLat: _prevLat,
      previousLng: _prevLng,
      previousAt: _prevAt,
      at: now,
    );
    if (validation.isFailure) return false;

    final result = await _ref.read(navigationRepositoryProvider).pingLocation(
          lat: lat,
          lng: lng,
          accuracyMeters: accuracyMeters,
          at: now,
        );
    if (result.isSuccess) {
      _prevLat = lat;
      _prevLng = lng;
      _prevAt = now;
    }
    return result.isSuccess;
  }
}
