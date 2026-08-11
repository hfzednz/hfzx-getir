import 'package:equatable/equatable.dart';

enum DeliveryLifecycleStatus {
  assigned,
  enRouteStore,
  atStore,
  pickedUp,
  enRouteCustomer,
  arrived,
  delivered,
  failed,
  cancelled;

  String get apiValue => switch (this) {
        DeliveryLifecycleStatus.assigned => 'assigned',
        DeliveryLifecycleStatus.enRouteStore => 'en_route_store',
        DeliveryLifecycleStatus.atStore => 'at_store',
        DeliveryLifecycleStatus.pickedUp => 'picked_up',
        DeliveryLifecycleStatus.enRouteCustomer => 'en_route_customer',
        DeliveryLifecycleStatus.arrived => 'arrived',
        DeliveryLifecycleStatus.delivered => 'delivered',
        DeliveryLifecycleStatus.failed => 'failed',
        DeliveryLifecycleStatus.cancelled => 'cancelled',
      };

  static DeliveryLifecycleStatus fromApi(String? value) {
    switch (value) {
      case 'en_route_store':
        return DeliveryLifecycleStatus.enRouteStore;
      case 'at_store':
        return DeliveryLifecycleStatus.atStore;
      case 'picked_up':
        return DeliveryLifecycleStatus.pickedUp;
      case 'en_route_customer':
        return DeliveryLifecycleStatus.enRouteCustomer;
      case 'arrived':
        return DeliveryLifecycleStatus.arrived;
      case 'delivered':
        return DeliveryLifecycleStatus.delivered;
      case 'failed':
        return DeliveryLifecycleStatus.failed;
      case 'cancelled':
        return DeliveryLifecycleStatus.cancelled;
      case 'assigned':
      default:
        return DeliveryLifecycleStatus.assigned;
    }
  }
}

class GeoPoint extends Equatable {
  const GeoPoint({required this.lat, required this.lng});

  final double lat;
  final double lng;

  factory GeoPoint.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const GeoPoint(lat: 0, lng: 0);
    return GeoPoint(
      lat: (json['lat'] as num?)?.toDouble() ?? 0,
      lng: (json['lng'] as num?)?.toDouble() ?? 0,
    );
  }

  @override
  List<Object?> get props => [lat, lng];
}

class DeliveryJob extends Equatable {
  const DeliveryJob({
    required this.id,
    required this.orderId,
    required this.status,
    required this.storeName,
    required this.storeLocation,
    required this.customerArea,
    required this.customerLocation,
    this.handoffToken,
    this.otpRequired = false,
    this.payoutMinor = 0,
    this.currency = 'TRY',
    this.batchId,
    this.notes,
  });

  final String id;
  final String orderId;
  final DeliveryLifecycleStatus status;
  final String storeName;
  final GeoPoint storeLocation;
  final String customerArea;
  final GeoPoint customerLocation;
  final String? handoffToken;
  final bool otpRequired;
  final int payoutMinor;
  final String currency;
  final String? batchId;
  final String? notes;

  factory DeliveryJob.fromJson(Map<String, dynamic> json) {
    return DeliveryJob(
      id: json['id']?.toString() ?? '',
      orderId: json['order_id']?.toString() ?? '',
      status: DeliveryLifecycleStatus.fromApi(json['status']?.toString()),
      storeName: json['store_name']?.toString() ?? '',
      storeLocation: GeoPoint.fromJson(
        json['store_location'] as Map<String, dynamic>? ??
            {
              'lat': json['store_lat'],
              'lng': json['store_lng'],
            },
      ),
      customerArea: json['customer_area']?.toString() ?? '',
      customerLocation: GeoPoint.fromJson(
        json['customer_location'] as Map<String, dynamic>? ??
            {
              'lat': json['customer_lat'],
              'lng': json['customer_lng'],
            },
      ),
      handoffToken: json['handoff_token']?.toString() ??
          json['warehouse_handoff']?.toString(),
      otpRequired: json['otp_required'] == true,
      payoutMinor: (json['payout_minor'] as num?)?.toInt() ?? 0,
      currency: json['currency']?.toString() ?? 'TRY',
      batchId: json['batch_id']?.toString(),
      notes: json['notes']?.toString(),
    );
  }

  @override
  List<Object?> get props => [
        id,
        orderId,
        status,
        storeName,
        storeLocation,
        customerArea,
        customerLocation,
        handoffToken,
        otpRequired,
        payoutMinor,
        currency,
        batchId,
        notes,
      ];
}
