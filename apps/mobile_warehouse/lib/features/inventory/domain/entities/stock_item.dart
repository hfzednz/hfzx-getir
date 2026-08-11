import 'package:equatable/equatable.dart';

class StockItem extends Equatable {
  const StockItem({
    required this.sku,
    required this.name,
    required this.onHand,
    this.bin,
    this.zone,
    this.reorderPoint = 0,
    this.expiryDate,
  });

  final String sku;
  final String name;
  final int onHand;
  final String? bin;
  final String? zone;
  final int reorderPoint;
  final DateTime? expiryDate;

  factory StockItem.fromJson(Map<String, dynamic> json) {
    return StockItem(
      sku: json['sku']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      onHand: (json['on_hand'] as num?)?.toInt() ?? 0,
      bin: json['bin']?.toString(),
      zone: json['zone']?.toString(),
      reorderPoint: (json['reorder_point'] as num?)?.toInt() ?? 0,
      expiryDate: DateTime.tryParse(json['expiry_date']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props => [sku, name, onHand, bin, zone, reorderPoint];
}

class CycleCountSession extends Equatable {
  const CycleCountSession({
    required this.id,
    required this.status,
    this.countedSkus = 0,
    this.totalSkus = 0,
  });

  final String id;
  final String status;
  final int countedSkus;
  final int totalSkus;

  factory CycleCountSession.fromJson(Map<String, dynamic> json) {
    return CycleCountSession(
      id: json['id']?.toString() ?? '',
      status: json['status']?.toString() ?? 'open',
      countedSkus: (json['counted_skus'] as num?)?.toInt() ?? 0,
      totalSkus: (json['total_skus'] as num?)?.toInt() ?? 0,
    );
  }

  @override
  List<Object?> get props => [id, status, countedSkus, totalSkus];
}
