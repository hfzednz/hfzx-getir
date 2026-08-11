import 'package:equatable/equatable.dart';

class ReturnTask extends Equatable {
  const ReturnTask({
    required this.id,
    required this.type,
    required this.status,
    this.reference,
    this.sku,
    this.qty = 0,
  });
  final String id;
  final String type; // customer | courier | supplier
  final String status;
  final String? reference;
  final String? sku;
  final int qty;
  factory ReturnTask.fromJson(Map<String, dynamic> json) => ReturnTask(
        id: json['id']?.toString() ?? '',
        type: json['type']?.toString() ?? 'customer',
        status: json['status']?.toString() ?? 'open',
        reference: json['reference']?.toString(),
        sku: json['sku']?.toString(),
        qty: (json['qty'] as num?)?.toInt() ?? 0,
      );
  @override
  List<Object?> get props => [id, type, status, reference, sku, qty];
}
