import 'package:equatable/equatable.dart';

import '../../../deliveries/domain/entities/delivery_job.dart';

class RoutePoint extends Equatable {
  const RoutePoint({required this.lat, required this.lng});

  final double lat;
  final double lng;

  factory RoutePoint.fromJson(Map<String, dynamic> json) => RoutePoint(
        lat: (json['lat'] as num?)?.toDouble() ?? 0,
        lng: (json['lng'] as num?)?.toDouble() ?? 0,
      );

  GeoPoint toGeo() => GeoPoint(lat: lat, lng: lng);

  @override
  List<Object?> get props => [lat, lng];
}

class DeliveryRoute extends Equatable {
  const DeliveryRoute({
    required this.deliveryId,
    required this.points,
    this.etaMinutes,
    this.distanceMeters,
    this.polylineEncoded,
  });

  final String deliveryId;
  final List<RoutePoint> points;
  final int? etaMinutes;
  final double? distanceMeters;
  final String? polylineEncoded;

  factory DeliveryRoute.fromJson(Map<String, dynamic> json) {
    final raw = json['points'] as List? ?? json['route_points'] as List? ?? [];
    return DeliveryRoute(
      deliveryId: json['delivery_id']?.toString() ?? json['id']?.toString() ?? '',
      points: raw
          .map((e) => RoutePoint.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
      etaMinutes: (json['eta_minutes'] as num?)?.toInt(),
      distanceMeters: (json['distance_meters'] as num?)?.toDouble(),
      polylineEncoded: json['polyline']?.toString(),
    );
  }

  @override
  List<Object?> get props =>
      [deliveryId, points, etaMinutes, distanceMeters, polylineEncoded];
}
