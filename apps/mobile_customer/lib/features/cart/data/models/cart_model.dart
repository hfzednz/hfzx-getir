import '../../domain/entities/cart_entity.dart';

class CartModel {
  const CartModel({required this.entity});

  final Cart entity;

  factory CartModel.fromJson(Map<String, dynamic> json) =>
      CartModel(entity: Cart.fromJson(json));

  Cart toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}
