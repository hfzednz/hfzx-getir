import 'package:equatable/equatable.dart';

enum HomeWidgetType {
  personalized,
  campaign,
  recommendation,
  trending,
  seasonal,
  recentlyViewed,
  continueShopping,
  brands,
  favoriteCategories,
  flashSale,
  countdown,
  banner,
}

HomeWidgetType homeWidgetTypeFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'personalized':
      return HomeWidgetType.personalized;
    case 'campaign':
      return HomeWidgetType.campaign;
    case 'recommendation':
      return HomeWidgetType.recommendation;
    case 'trending':
      return HomeWidgetType.trending;
    case 'seasonal':
      return HomeWidgetType.seasonal;
    case 'recently_viewed':
      return HomeWidgetType.recentlyViewed;
    case 'continue_shopping':
      return HomeWidgetType.continueShopping;
    case 'brands':
      return HomeWidgetType.brands;
    case 'favorite_categories':
      return HomeWidgetType.favoriteCategories;
    case 'flash_sale':
      return HomeWidgetType.flashSale;
    case 'countdown':
      return HomeWidgetType.countdown;
    case 'banner':
      return HomeWidgetType.banner;
    default:
      return HomeWidgetType.recommendation;
  }
}

String homeWidgetTypeToJson(HomeWidgetType type) => switch (type) {
      HomeWidgetType.personalized => 'personalized',
      HomeWidgetType.campaign => 'campaign',
      HomeWidgetType.recommendation => 'recommendation',
      HomeWidgetType.trending => 'trending',
      HomeWidgetType.seasonal => 'seasonal',
      HomeWidgetType.recentlyViewed => 'recently_viewed',
      HomeWidgetType.continueShopping => 'continue_shopping',
      HomeWidgetType.brands => 'brands',
      HomeWidgetType.favoriteCategories => 'favorite_categories',
      HomeWidgetType.flashSale => 'flash_sale',
      HomeWidgetType.countdown => 'countdown',
      HomeWidgetType.banner => 'banner',
    };

/// Lightweight product tile used on home rails and category grids.
class HomeProduct extends Equatable {
  const HomeProduct({
    required this.id,
    required this.title,
    required this.priceMinor,
    this.currency = 'TRY',
    this.imageUrl,
    this.unitMeta,
    this.compareAtMinor,
    this.deepLink,
  });

  final String id;
  final String title;
  final int priceMinor;
  final String currency;
  final String? imageUrl;
  final String? unitMeta;
  final int? compareAtMinor;
  final String? deepLink;

  factory HomeProduct.fromJson(Map<String, dynamic> json) => HomeProduct(
        id: json['id']?.toString() ??
            json['productId']?.toString() ??
            json['sku']?.toString() ??
            '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        priceMinor: (json['price_minor'] as num?)?.toInt() ??
            (json['priceMinor'] as num?)?.toInt() ??
            (json['unit_price_minor'] as num?)?.toInt() ??
            0,
        currency: json['currency']?.toString() ?? 'TRY',
        imageUrl: json['image_url']?.toString() ?? json['imageUrl']?.toString(),
        unitMeta: json['unit_meta']?.toString(),
        compareAtMinor: (json['compare_at_minor'] as num?)?.toInt(),
        deepLink: json['deep_link']?.toString() ?? json['deepLink']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        'price_minor': priceMinor,
        'currency': currency,
        if (imageUrl != null) 'image_url': imageUrl,
        if (unitMeta != null) 'unit_meta': unitMeta,
        if (compareAtMinor != null) 'compare_at_minor': compareAtMinor,
        if (deepLink != null) 'deep_link': deepLink,
      };

  @override
  List<Object?> get props =>
      [id, title, priceMinor, currency, imageUrl, unitMeta, compareAtMinor, deepLink];
}

class HomeWidgetConfig extends Equatable {
  const HomeWidgetConfig({
    required this.id,
    required this.type,
    this.title = '',
    this.payload = const {},
    this.items = const [],
  });

  final String id;
  final HomeWidgetType type;
  final String title;
  final Map<String, dynamic> payload;
  final List<HomeProduct> items;

  factory HomeWidgetConfig.fromJson(Map<String, dynamic> json) => HomeWidgetConfig(
        id: json['id']?.toString() ?? json['type']?.toString() ?? '',
        type: homeWidgetTypeFromJson(json['type']?.toString()),
        title: json['title']?.toString() ?? '',
        payload: Map<String, dynamic>.from(json['payload'] as Map? ?? {}),
        items: (json['items'] as List<dynamic>? ?? json['products'] as List<dynamic>? ?? [])
            .map((e) => HomeProduct.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': homeWidgetTypeToJson(type),
        'title': title,
        'payload': payload,
        'items': items.map((e) => e.toJson()).toList(),
      };

  @override
  List<Object?> get props => [id, type, title, payload, items];
}

class HomeFeed extends Equatable {
  const HomeFeed({this.widgets = const [], this.serviceable = true});

  final List<HomeWidgetConfig> widgets;
  final bool serviceable;

  factory HomeFeed.fromJson(Map<String, dynamic> json) {
    final serviceable = json['serviceable'] != false;
    final widgetsRaw = json['widgets'] as List<dynamic>? ?? [];
    if (widgetsRaw.isNotEmpty) {
      return HomeFeed(
        serviceable: serviceable,
        widgets: widgetsRaw
            .map((e) => HomeWidgetConfig.fromJson(e as Map<String, dynamic>))
            .toList(),
      );
    }
    final productsRaw = json['Products'] ?? json['products'] ?? json['hits'];
    final products = <HomeProduct>[];
    if (productsRaw is List) {
      for (final item in productsRaw) {
        if (item is Map) {
          final m = Map<String, dynamic>.from(item);
          products.add(
            HomeProduct(
              id: (m['productId'] ?? m['id'] ?? m['sku'] ?? '').toString(),
              title: (m['name'] ?? m['title'] ?? m['sku'] ?? 'Item').toString(),
              priceMinor: (m['priceMinor'] as num?)?.toInt() ??
                  (m['price_minor'] as num?)?.toInt() ??
                  0,
            ),
          );
        }
      }
    }
    if (products.isEmpty) {
      return HomeFeed(serviceable: serviceable);
    }
    return HomeFeed(
      serviceable: serviceable,
      widgets: [
        HomeWidgetConfig(
          id: 'bff-home',
          type: HomeWidgetType.trending,
          title: 'Popular',
          items: products,
        ),
      ],
    );
  }

  Map<String, dynamic> toJson() => {
        'widgets': widgets.map((e) => e.toJson()).toList(),
        'serviceable': serviceable,
      };

  @override
  List<Object?> get props => [widgets, serviceable];
}
