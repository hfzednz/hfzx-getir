import 'package:equatable/equatable.dart';

enum HandoffStatus {
  waitingCourier,
  courierArrived,
  verifying,
  handedOff,
  failedPickup;

  static HandoffStatus fromString(String? raw) {
    final v = (raw ?? '').toLowerCase().replaceAll('-', '_');
    return switch (v) {
      'waiting_courier' || 'waitingcourier' => HandoffStatus.waitingCourier,
      'courier_arrived' || 'courierarrived' => HandoffStatus.courierArrived,
      'verifying' => HandoffStatus.verifying,
      'handed_off' || 'handedoff' => HandoffStatus.handedOff,
      'failed_pickup' || 'failedpickup' => HandoffStatus.failedPickup,
      _ => HandoffStatus.waitingCourier,
    };
  }

  String get wireName => switch (this) {
        HandoffStatus.waitingCourier => 'waiting_courier',
        HandoffStatus.courierArrived => 'courier_arrived',
        HandoffStatus.handedOff => 'handed_off',
        HandoffStatus.failedPickup => 'failed_pickup',
        _ => name,
      };
}

class HandoffTask extends Equatable {
  const HandoffTask({
    required this.id,
    required this.orderId,
    required this.status,
    required this.handoffToken,
    this.courierId,
    this.courierName,
    this.bagCount = 1,
    this.arrivedAt,
  });

  final String id;
  final String orderId;
  final HandoffStatus status;
  final String handoffToken;
  final String? courierId;
  final String? courierName;
  final int bagCount;
  final DateTime? arrivedAt;

  factory HandoffTask.fromJson(Map<String, dynamic> json) {
    return HandoffTask(
      id: json['id']?.toString() ?? '',
      orderId: json['order_id']?.toString() ?? '',
      status: HandoffStatus.fromString(json['status']?.toString()),
      handoffToken: json['handoff_token']?.toString() ??
          json['token']?.toString() ??
          '',
      courierId: json['courier_id']?.toString(),
      courierName: json['courier_name']?.toString(),
      bagCount: (json['bag_count'] as num?)?.toInt() ?? 1,
      arrivedAt: DateTime.tryParse(json['arrived_at']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props =>
      [id, orderId, status, handoffToken, courierId, bagCount];
}
