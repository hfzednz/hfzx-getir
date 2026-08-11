import 'package:nexora_core/nexora_core.dart';

import '../entities/route_entity.dart';

abstract class NavigationRepository {
  Future<Result<DeliveryRoute>> fetchRoute(String deliveryId);
  Future<Result<void>> pingLocation({
    required double lat,
    required double lng,
    double? accuracyMeters,
    DateTime? at,
  });
}
