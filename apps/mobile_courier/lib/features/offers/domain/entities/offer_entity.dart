import 'package:equatable/equatable.dart';

class Offer extends Equatable {
  const Offer({
    required this.id,
    required this.orderId,
    required this.storeName,
    required this.storeLat,
    required this.storeLng,
    required this.customerArea,
    required this.payoutMinor,
    required this.currency,
    required this.expiresAt,
    this.priority = 0,
    this.batchId,
  });

  final String id;
  final String orderId;
  final String storeName;
  final double storeLat;
  final double storeLng;
  final String customerArea;
  final int payoutMinor;
  final String currency;
  final DateTime expiresAt;
  final int priority;
  final String? batchId;

  bool get isExpired => expiresAt.isBefore(DateTime.now());

  factory Offer.fromJson(Map<String, dynamic> json) {
    return Offer(
      id: json['id']?.toString() ?? '',
      orderId: json['order_id']?.toString() ?? '',
      storeName: json['store_name']?.toString() ?? '',
      storeLat: (json['store_lat'] as num?)?.toDouble() ?? 0,
      storeLng: (json['store_lng'] as num?)?.toDouble() ?? 0,
      customerArea: json['customer_area']?.toString() ?? '',
      payoutMinor: (json['payout_minor'] as num?)?.toInt() ?? 0,
      currency: json['currency']?.toString() ?? 'TRY',
      expiresAt: DateTime.tryParse(json['expires_at']?.toString() ?? '') ??
          DateTime.now().add(const Duration(minutes: 1)),
      priority: (json['priority'] as num?)?.toInt() ?? 0,
      batchId: json['batch_id']?.toString(),
    );
  }

  @override
  List<Object?> get props => [
        id,
        orderId,
        storeName,
        storeLat,
        storeLng,
        customerArea,
        payoutMinor,
        currency,
        expiresAt,
        priority,
        batchId,
      ];
}
