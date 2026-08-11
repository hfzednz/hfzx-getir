import '../../domain/entities/orders_entity.dart';

class OrderModel {
  const OrderModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory OrderModel.fromJson(Map<String, dynamic> json) => OrderModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  Order toEntity() => Order.fromJson(raw);
}
