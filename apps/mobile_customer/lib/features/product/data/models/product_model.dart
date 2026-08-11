import '../../domain/entities/product_entity.dart';

class ProductModel {
  const ProductModel({required this.entity});

  final Product entity;

  factory ProductModel.fromJson(Map<String, dynamic> json) =>
      ProductModel(entity: Product.fromJson(json));

  Product toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}

class ProductPriceHistoryModel {
  const ProductPriceHistoryModel({required this.points});

  final List<ProductPricePoint> points;

  factory ProductPriceHistoryModel.fromJson(Map<String, dynamic> json) =>
      ProductPriceHistoryModel(
        points: (json['points'] as List<dynamic>? ?? json['history'] as List<dynamic>? ?? [])
            .map((e) => ProductPricePoint.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  List<ProductPricePoint> toEntity() => points;
}
