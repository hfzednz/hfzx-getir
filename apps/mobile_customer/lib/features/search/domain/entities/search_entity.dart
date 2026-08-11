import 'package:equatable/equatable.dart';

import '../../../product/domain/entities/product_entity.dart';

enum SearchSort { relevance, priceAsc, priceDesc, newest, rating }

SearchSort searchSortFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'price_asc':
      return SearchSort.priceAsc;
    case 'price_desc':
      return SearchSort.priceDesc;
    case 'newest':
      return SearchSort.newest;
    case 'rating':
      return SearchSort.rating;
    default:
      return SearchSort.relevance;
  }
}

String searchSortToJson(SearchSort sort) => switch (sort) {
      SearchSort.relevance => 'relevance',
      SearchSort.priceAsc => 'price_asc',
      SearchSort.priceDesc => 'price_desc',
      SearchSort.newest => 'newest',
      SearchSort.rating => 'rating',
    };

class SearchFilters extends Equatable {
  const SearchFilters({
    this.brands = const [],
    this.priceMinMinor,
    this.priceMaxMinor,
    this.availability,
    this.dietary = const [],
    this.sort = SearchSort.relevance,
    this.categoryId,
  });

  final List<String> brands;
  final int? priceMinMinor;
  final int? priceMaxMinor;
  final ProductStockStatus? availability;
  final List<String> dietary;
  final SearchSort sort;
  final String? categoryId;

  Map<String, dynamic> toQueryParams() => {
        if (brands.isNotEmpty) 'brands': brands.join(','),
        if (priceMinMinor != null) 'price_min_minor': priceMinMinor,
        if (priceMaxMinor != null) 'price_max_minor': priceMaxMinor,
        if (availability != null) 'availability': productStockStatusToJson(availability!),
        if (dietary.isNotEmpty) 'dietary': dietary.join(','),
        'sort': searchSortToJson(sort),
        if (categoryId != null) 'category_id': categoryId,
      };

  factory SearchFilters.fromJson(Map<String, dynamic> json) => SearchFilters(
        brands: (json['brands'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList(),
        priceMinMinor: (json['price_min_minor'] as num?)?.toInt(),
        priceMaxMinor: (json['price_max_minor'] as num?)?.toInt(),
        availability: json['availability'] != null
            ? productStockStatusFromJson(json['availability']?.toString())
            : null,
        dietary: (json['dietary'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList(),
        sort: searchSortFromJson(json['sort']?.toString()),
        categoryId: json['category_id']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'brands': brands,
        if (priceMinMinor != null) 'price_min_minor': priceMinMinor,
        if (priceMaxMinor != null) 'price_max_minor': priceMaxMinor,
        if (availability != null) 'availability': productStockStatusToJson(availability!),
        'dietary': dietary,
        'sort': searchSortToJson(sort),
        if (categoryId != null) 'category_id': categoryId,
      };

  SearchFilters copyWith({
    List<String>? brands,
    int? priceMinMinor,
    int? priceMaxMinor,
    ProductStockStatus? availability,
    List<String>? dietary,
    SearchSort? sort,
    String? categoryId,
    bool clearAvailability = false,
  }) =>
      SearchFilters(
        brands: brands ?? this.brands,
        priceMinMinor: priceMinMinor ?? this.priceMinMinor,
        priceMaxMinor: priceMaxMinor ?? this.priceMaxMinor,
        availability: clearAvailability ? null : (availability ?? this.availability),
        dietary: dietary ?? this.dietary,
        sort: sort ?? this.sort,
        categoryId: categoryId ?? this.categoryId,
      );

  @override
  List<Object?> get props =>
      [brands, priceMinMinor, priceMaxMinor, availability, dietary, sort, categoryId];
}

class SearchSuggestion extends Equatable {
  const SearchSuggestion({
    required this.text,
    this.type = 'query',
    this.productId,
  });

  final String text;
  final String type;
  final String? productId;

  factory SearchSuggestion.fromJson(Map<String, dynamic> json) => SearchSuggestion(
        text: json['text']?.toString() ?? json['query']?.toString() ?? '',
        type: json['type']?.toString() ?? 'query',
        productId: json['product_id']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'text': text,
        'type': type,
        if (productId != null) 'product_id': productId,
      };

  @override
  List<Object?> get props => [text, type, productId];
}

class SearchResult extends Equatable {
  const SearchResult({
    required this.id,
    this.query = '',
    this.items = const [],
    this.suggestions = const [],
    this.appliedFilters = const SearchFilters(),
    this.totalCount = 0,
    this.nextCursor,
  });

  final String id;
  final String query;
  final List<ProductSummary> items;
  final List<SearchSuggestion> suggestions;
  final SearchFilters appliedFilters;
  final int totalCount;
  final String? nextCursor;

  factory SearchResult.fromJson(Map<String, dynamic> json) => SearchResult(
        id: json['id']?.toString() ?? json['query']?.toString() ?? '',
        query: json['query']?.toString() ?? '',
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => ProductSummary.fromJson(e as Map<String, dynamic>))
            .toList(),
        suggestions: (json['suggestions'] as List<dynamic>? ?? [])
            .map((e) => SearchSuggestion.fromJson(e as Map<String, dynamic>))
            .toList(),
        appliedFilters: json['filters'] != null
            ? SearchFilters.fromJson(json['filters'] as Map<String, dynamic>)
            : const SearchFilters(),
        totalCount: (json['total_count'] as num?)?.toInt() ?? 0,
        nextCursor: json['next_cursor']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'query': query,
        'items': items.map((e) => e.toJson()).toList(),
        'suggestions': suggestions.map((e) => e.toJson()).toList(),
        'filters': appliedFilters.toJson(),
        'total_count': totalCount,
        if (nextCursor != null) 'next_cursor': nextCursor,
      };

  @override
  List<Object?> get props =>
      [id, query, items, suggestions, appliedFilters, totalCount, nextCursor];
}
