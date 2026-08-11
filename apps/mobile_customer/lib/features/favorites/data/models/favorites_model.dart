import '../../domain/entities/favorites_entity.dart';

class FavoriteEntryModel {
  const FavoriteEntryModel({required this.raw});
  final Map<String, dynamic> raw;
  factory FavoriteEntryModel.fromJson(Map<String, dynamic> json) => FavoriteEntryModel(raw: json);
  FavoriteEntry toEntity() => FavoriteEntry.fromJson(raw);
}

class FavoriteItemModel {
  const FavoriteItemModel({required this.raw});
  final Map<String, dynamic> raw;
  factory FavoriteItemModel.fromJson(Map<String, dynamic> json) => FavoriteItemModel(raw: json);
  FavoriteItem toEntity() => FavoriteItem.fromJson(raw);
}
