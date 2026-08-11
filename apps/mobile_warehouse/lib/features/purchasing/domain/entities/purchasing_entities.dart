import 'package:equatable/equatable.dart';

class Supplier extends Equatable {
  const Supplier({required this.id, required this.name, this.contact});
  final String id;
  final String name;
  final String? contact;
  factory Supplier.fromJson(Map<String, dynamic> json) => Supplier(
        id: json['id']?.toString() ?? '',
        name: json['name']?.toString() ?? '',
        contact: json['contact']?.toString(),
      );
  @override
  List<Object?> get props => [id, name, contact];
}

class PurchaseOrder extends Equatable {
  const PurchaseOrder({
    required this.id,
    required this.supplierId,
    required this.status,
    this.supplierName,
    this.lineCount = 0,
  });
  final String id;
  final String supplierId;
  final String status;
  final String? supplierName;
  final int lineCount;
  factory PurchaseOrder.fromJson(Map<String, dynamic> json) => PurchaseOrder(
        id: json['id']?.toString() ?? '',
        supplierId: json['supplier_id']?.toString() ?? '',
        status: json['status']?.toString() ?? 'open',
        supplierName: json['supplier_name']?.toString(),
        lineCount: (json['line_count'] as num?)?.toInt() ?? 0,
      );
  @override
  List<Object?> get props => [id, supplierId, status, supplierName, lineCount];
}
