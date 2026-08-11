import '../../domain/entities/search_entity.dart';

class SearchResultModel {
  const SearchResultModel({required this.entity});

  final SearchResult entity;

  factory SearchResultModel.fromJson(Map<String, dynamic> json) =>
      SearchResultModel(entity: SearchResult.fromJson(json));

  SearchResult toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}
