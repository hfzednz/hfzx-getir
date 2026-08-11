import 'package:equatable/equatable.dart';

enum CategorySort { nameAsc, nameDesc, productCount, featured }

CategorySort categorySortFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'name_desc':
      return CategorySort.nameDesc;
    case 'product_count':
      return CategorySort.productCount;
    case 'featured':
      return CategorySort.featured;
    default:
      return CategorySort.nameAsc;
  }
}

String categorySortToJson(CategorySort sort) => switch (sort) {
      CategorySort.nameAsc => 'name_asc',
      CategorySort.nameDesc => 'name_desc',
      CategorySort.productCount => 'product_count',
      CategorySort.featured => 'featured',
    };

class CategoryFilter extends Equatable {
  const CategoryFilter({
    this.brands = const [],
    this.priceMinMinor,
    this.priceMaxMinor,
    this.dietary = const [],
    this.inStockOnly = false,
    this.sort = CategorySort.featured,
  });

  final List<String> brands;
  final int? priceMinMinor;
  final int? priceMaxMinor;
  final List<String> dietary;
  final bool inStockOnly;
  final CategorySort sort;

  Map<String, dynamic> toQueryParams() => {
        if (brands.isNotEmpty) 'brands': brands.join(','),
        if (priceMinMinor != null) 'price_min_minor': priceMinMinor,
        if (priceMaxMinor != null) 'price_max_minor': priceMaxMinor,
        if (dietary.isNotEmpty) 'dietary': dietary.join(','),
        if (inStockOnly) 'in_stock_only': true,
        'sort': categorySortToJson(sort),
      };

  factory CategoryFilter.fromJson(Map<String, dynamic> json) => CategoryFilter(
        brands: (json['brands'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList(),
        priceMinMinor: (json['price_min_minor'] as num?)?.toInt(),
        priceMaxMinor: (json['price_max_minor'] as num?)?.toInt(),
        dietary: (json['dietary'] as List<dynamic>? ?? [])
            .map((e) => e.toString())
            .toList(),
        inStockOnly: json['in_stock_only'] == true,
        sort: categorySortFromJson(json['sort']?.toString()),
      );

  Map<String, dynamic> toJson() => {
        'brands': brands,
        if (priceMinMinor != null) 'price_min_minor': priceMinMinor,
        if (priceMaxMinor != null) 'price_max_minor': priceMaxMinor,
        'dietary': dietary,
        'in_stock_only': inStockOnly,
        'sort': categorySortToJson(sort),
      };

  CategoryFilter copyWith({
    List<String>? brands,
    int? priceMinMinor,
    int? priceMaxMinor,
    List<String>? dietary,
    bool? inStockOnly,
    CategorySort? sort,
  }) =>
      CategoryFilter(
        brands: brands ?? this.brands,
        priceMinMinor: priceMinMinor ?? this.priceMinMinor,
        priceMaxMinor: priceMaxMinor ?? this.priceMaxMinor,
        dietary: dietary ?? this.dietary,
        inStockOnly: inStockOnly ?? this.inStockOnly,
        sort: sort ?? this.sort,
      );

  @override
  List<Object?> get props =>
      [brands, priceMinMinor, priceMaxMinor, dietary, inStockOnly, sort];
}

class Category extends Equatable {
  const Category({
    required this.id,
    this.title = '',
    this.slug,
    this.imageUrl,
    this.iconUrl,
    this.productCount,
    this.parentId,
    this.children = const [],
    this.depth = 0,
  });

  final String id;
  final String title;
  final String? slug;
  final String? imageUrl;
  final String? iconUrl;
  final int? productCount;
  final String? parentId;
  final List<Category> children;
  final int depth;

  bool get hasChildren => children.isNotEmpty;

  factory Category.fromJson(Map<String, dynamic> json) {
    List<Category> parseChildren(dynamic raw, int depth) {
      if (raw is! List) return const [];
      return raw
          .map((e) => Category.fromJson(e as Map<String, dynamic>).copyWithDepth(depth))
          .toList();
    }

    return Category(
      id: json['id']?.toString() ?? '',
      title: json['title']?.toString() ?? json['name']?.toString() ?? '',
      slug: json['slug']?.toString(),
      imageUrl: json['image_url']?.toString(),
      iconUrl: json['icon_url']?.toString(),
      productCount: (json['product_count'] as num?)?.toInt(),
      parentId: json['parent_id']?.toString(),
      children: parseChildren(json['children'], 1),
    );
  }

  Category copyWithDepth(int depth) => Category(
        id: id,
        title: title,
        slug: slug,
        imageUrl: imageUrl,
        iconUrl: iconUrl,
        productCount: productCount,
        parentId: parentId,
        children: children
            .map((c) => c.copyWithDepth(depth + 1))
            .toList(),
        depth: depth,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        if (slug != null) 'slug': slug,
        if (imageUrl != null) 'image_url': imageUrl,
        if (iconUrl != null) 'icon_url': iconUrl,
        if (productCount != null) 'product_count': productCount,
        if (parentId != null) 'parent_id': parentId,
        'children': children.map((e) => e.toJson()).toList(),
      };

  @override
  List<Object?> get props =>
      [id, title, slug, imageUrl, iconUrl, productCount, parentId, children, depth];
}
