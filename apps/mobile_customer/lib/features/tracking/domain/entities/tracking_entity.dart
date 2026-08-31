import 'package:equatable/equatable.dart';

class LatLngPoint extends Equatable {
  const LatLngPoint({required this.lat, required this.lng});

  final double lat;
  final double lng;

  factory LatLngPoint.fromJson(Map<String, dynamic> json) => LatLngPoint(
        lat: (json['lat'] as num?)?.toDouble() ??
            (json['latitude'] as num?)?.toDouble() ??
            0,
        lng: (json['lng'] as num?)?.toDouble() ??
            (json['lon'] as num?)?.toDouble() ??
            (json['longitude'] as num?)?.toDouble() ??
            0,
      );

  Map<String, dynamic> toJson() => {'lat': lat, 'lng': lng};

  @override
  List<Object?> get props => [lat, lng];
}

class TrackingStep extends Equatable {
  const TrackingStep({
    required this.title,
    this.subtitle,
    this.state = 'upcoming',
  });

  final String title;
  final String? subtitle;
  final String state;

  factory TrackingStep.fromJson(Map<String, dynamic> json) => TrackingStep(
        title: json['title']?.toString() ??
            json['label']?.toString() ??
            '',
        subtitle: json['subtitle']?.toString(),
        state: json['state']?.toString() ?? 'upcoming',
      );

  Map<String, dynamic> toJson() => {
        'title': title,
        if (subtitle != null) 'subtitle': subtitle,
        'state': state,
      };

  @override
  List<Object?> get props => [title, subtitle, state];
}

class TrackingSnapshot extends Equatable {
  const TrackingSnapshot({
    required this.orderId,
    required this.status,
    this.etaMinutes,
    this.etaMin,
    this.etaMax,
    this.courierName,
    this.courierPhone,
    this.courierLat,
    this.courierLng,
    this.storeLat,
    this.storeLng,
    this.destLat,
    this.destLng,
    this.steps = const [],
    this.routePoints = const [],
    this.canCall = false,
    this.canChat = true,
    this.courierChatUrl,
  });

  final String orderId;
  final String status;
  final int? etaMinutes;
  final int? etaMin;
  final int? etaMax;
  final String? courierName;
  final String? courierPhone;
  final double? courierLat;
  final double? courierLng;
  final double? storeLat;
  final double? storeLng;
  final double? destLat;
  final double? destLng;
  final List<TrackingStep> steps;
  final List<LatLngPoint> routePoints;
  final bool canCall;
  final bool canChat;
  final String? courierChatUrl;

  bool get hasMapCoordinates =>
      (storeLat != null && storeLng != null) ||
      (destLat != null && destLng != null) ||
      (courierLat != null && courierLng != null) ||
      routePoints.length >= 2;

  factory TrackingSnapshot.fromJson(Map<String, dynamic> json) {
    final courier = json['courier'] as Map<String, dynamic>?;
    final store = json['store'] as Map<String, dynamic>?;
    final destination = json['destination'] as Map<String, dynamic>? ??
        json['dest'] as Map<String, dynamic>?;

    double? coord(dynamic value) => (value as num?)?.toDouble();

    return TrackingSnapshot(
      orderId: json['order_id']?.toString() ??
          json['orderId']?.toString() ??
          json['id']?.toString() ??
          '',
      status: json['status']?.toString() ?? 'unknown',
      etaMinutes: () {
        final minutes = (json['eta_minutes'] as num?)?.toInt();
        if (minutes != null) return minutes;
        final seconds = (json['etaSeconds'] as num?)?.toInt();
        if (seconds != null && seconds > 0) return (seconds / 60).round();
        return null;
      }(),
      etaMin: (json['eta_min'] as num?)?.toInt(),
      etaMax: (json['eta_max'] as num?)?.toInt(),
      courierName: json['courier_name']?.toString() ??
          json['courierId']?.toString() ??
          courier?['name']?.toString(),
      courierPhone: json['courier_phone']?.toString() ??
          courier?['phone']?.toString(),
      courierLat: coord(json['courier_lat']) ??
          coord(json['lat']) ??
          coord(courier?['lat']),
      courierLng: coord(json['courier_lng']) ??
          coord(json['lng']) ??
          coord(courier?['lng']),
      storeLat: coord(json['store_lat']) ?? coord(store?['lat']),
      storeLng: coord(json['store_lng']) ?? coord(store?['lng']),
      destLat: coord(json['dest_lat']) ?? coord(destination?['lat']),
      destLng: coord(json['dest_lng']) ?? coord(destination?['lng']),
      steps: (json['steps'] as List<dynamic>? ?? [])
          .map((e) => TrackingStep.fromJson(e as Map<String, dynamic>))
          .toList(),
      routePoints: _parseRoutePoints(json),
      canCall: json['can_call'] == true ||
          (json['courier_phone'] != null || courier?['phone'] != null),
      canChat: json['can_chat'] != false,
      courierChatUrl: json['courier_chat_url']?.toString() ??
          json['courierChatUrl']?.toString() ??
          courier?['chat_url']?.toString() ??
          courier?['chatUrl']?.toString(),
    );
  }

  static List<LatLngPoint> _parseRoutePoints(Map<String, dynamic> json) {
    final raw = json['route'] ?? json['polyline'] ?? json['route_points'];
    if (raw is! List) return const [];
    final points = <LatLngPoint>[];
    for (final entry in raw) {
      if (entry is Map<String, dynamic>) {
        points.add(LatLngPoint.fromJson(entry));
      } else if (entry is Map) {
        points.add(LatLngPoint.fromJson(Map<String, dynamic>.from(entry)));
      } else if (entry is List && entry.length >= 2) {
        final lat = (entry[0] as num?)?.toDouble();
        final lng = (entry[1] as num?)?.toDouble();
        if (lat != null && lng != null) {
          points.add(LatLngPoint(lat: lat, lng: lng));
        }
      }
    }
    return points;
  }

  Map<String, dynamic> toJson() => {
        'order_id': orderId,
        'status': status,
        if (etaMinutes != null) 'eta_minutes': etaMinutes,
        if (etaMin != null) 'eta_min': etaMin,
        if (etaMax != null) 'eta_max': etaMax,
        if (courierName != null) 'courier_name': courierName,
        if (courierPhone != null) 'courier_phone': courierPhone,
        if (courierLat != null) 'courier_lat': courierLat,
        if (courierLng != null) 'courier_lng': courierLng,
        if (storeLat != null) 'store_lat': storeLat,
        if (storeLng != null) 'store_lng': storeLng,
        if (destLat != null) 'dest_lat': destLat,
        if (destLng != null) 'dest_lng': destLng,
        'steps': steps.map((s) => s.toJson()).toList(),
        if (routePoints.isNotEmpty)
          'route': routePoints.map((p) => p.toJson()).toList(),
        'can_call': canCall,
        'can_chat': canChat,
        if (courierChatUrl != null) 'courier_chat_url': courierChatUrl,
      };

  @override
  List<Object?> get props => [
        orderId,
        status,
        etaMinutes,
        etaMin,
        etaMax,
        courierName,
        courierPhone,
        courierLat,
        courierLng,
        storeLat,
        storeLng,
        destLat,
        destLng,
        steps,
        routePoints,
        canCall,
        canChat,
        courierChatUrl,
      ];
}
