import 'package:equatable/equatable.dart';

enum TransferStatus { draft, pending, approved, rejected, completed;

  static TransferStatus fromString(String? raw) {
    final v = (raw ?? '').toLowerCase();
    return switch (v) {
      'pending' => TransferStatus.pending,
      'approved' => TransferStatus.approved,
      'rejected' => TransferStatus.rejected,
      'completed' => TransferStatus.completed,
      _ => TransferStatus.draft,
    };
  }
}

class WarehouseTransfer extends Equatable {
  const WarehouseTransfer({
    required this.id,
    required this.type,
    required this.status,
    required this.fromLocation,
    required this.toLocation,
    this.sku,
    this.qty = 0,
  });

  final String id;
  final String type; // shelf | warehouse
  final TransferStatus status;
  final String fromLocation;
  final String toLocation;
  final String? sku;
  final int qty;

  factory WarehouseTransfer.fromJson(Map<String, dynamic> json) {
    return WarehouseTransfer(
      id: json['id']?.toString() ?? '',
      type: json['type']?.toString() ?? 'shelf',
      status: TransferStatus.fromString(json['status']?.toString()),
      fromLocation: json['from_location']?.toString() ?? '',
      toLocation: json['to_location']?.toString() ?? '',
      sku: json['sku']?.toString(),
      qty: (json['qty'] as num?)?.toInt() ?? 0,
    );
  }

  @override
  List<Object?> get props => [id, type, status, fromLocation, toLocation, sku, qty];
}
