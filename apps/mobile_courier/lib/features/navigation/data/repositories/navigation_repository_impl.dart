import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/route_entity.dart';
import '../../domain/repositories/navigation_repository.dart';
import '../datasources/navigation_remote_datasource.dart';

class NavigationRepositoryImpl implements NavigationRepository {
  NavigationRepositoryImpl(this._remote);
  final NavigationRemoteDataSource _remote;

  @override
  Future<Result<DeliveryRoute>> fetchRoute(String deliveryId) =>
      _remote.fetchRoute(deliveryId);

  @override
  Future<Result<void>> pingLocation({
    required double lat,
    required double lng,
    double? accuracyMeters,
    DateTime? at,
  }) =>
      _remote.pingLocation(
        lat: lat,
        lng: lng,
        accuracyMeters: accuracyMeters,
        at: at,
      );
}
