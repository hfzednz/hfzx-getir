import '../../domain/entities/categories_entity.dart';

class CategoryModel {
  const CategoryModel({required this.entity});

  final Category entity;

  factory CategoryModel.fromJson(Map<String, dynamic> json) =>
      CategoryModel(entity: Category.fromJson(json));

  Category toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}
