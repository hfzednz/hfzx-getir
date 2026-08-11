import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/route_entity.dart';

class NavigationRemoteDataSource {
  const NavigationRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<DeliveryRoute>> fetchRoute(String deliveryId) {
    return _client.get<DeliveryRoute>(
      '/courier/deliveries/$deliveryId/route',
      parser: (json) =>
          DeliveryRoute.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<void>> pingLocation({
    required double lat,
    required double lng,
    double? accuracyMeters,
    DateTime? at,
  }) {
    return _client.post<void>(
      '/courier/location',
      data: {
        'lat': lat,
        'lng': lng,
        if (accuracyMeters != null) 'accuracy_meters': accuracyMeters,
        'at': (at ?? DateTime.now()).toUtc().toIso8601String(),
      },
      parser: (_) {},
    );
  }
}
