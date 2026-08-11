import 'package:equatable/equatable.dart';

class ExpiryItem extends Equatable {
  const ExpiryItem({
    required this.sku,
    required this.name,
    required this.qty,
    required this.expiresAt,
    this.bin,
    this.fefoHint,
  });

  final String sku;
  final String name;
  final int qty;
  final DateTime expiresAt;
  final String? bin;
  final String? fefoHint;

  factory ExpiryItem.fromJson(Map<String, dynamic> json) {
    return ExpiryItem(
      sku: json['sku']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      qty: (json['qty'] as num?)?.toInt() ?? 0,
      expiresAt: DateTime.tryParse(json['expires_at']?.toString() ?? '') ?? DateTime.now(),
      bin: json['bin']?.toString(),
      fefoHint: json['fefo_hint']?.toString(),
    );
  }

  @override
  List<Object?> get props => [sku, name, qty, expiresAt, bin];
}
