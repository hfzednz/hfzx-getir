import 'package:equatable/equatable.dart';
import 'package:nexora_design/nexora_design.dart';

enum ProductStockStatus { inStock, low, outOfStock }

ProductStockStatus productStockStatusFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'low':
    case 'low_stock':
      return ProductStockStatus.low;
    case 'out':
    case 'out_of_stock':
      return ProductStockStatus.outOfStock;
    default:
      return ProductStockStatus.inStock;
  }
}

String productStockStatusToJson(ProductStockStatus status) => switch (status) {
      ProductStockStatus.inStock => 'in_stock',
      ProductStockStatus.low => 'low',
      ProductStockStatus.outOfStock => 'out',
    };

NxStockStatus toNxStockStatus(ProductStockStatus status) => switch (status) {
      ProductStockStatus.inStock => NxStockStatus.inStock,
      ProductStockStatus.low => NxStockStatus.low,
      ProductStockStatus.outOfStock => NxStockStatus.outOfStock,
    };

class ProductMedia extends Equatable {
  const ProductMedia({required this.url, this.type = 'image', this.alt});

  final String url;
  final String type;
  final String? alt;

  factory ProductMedia.fromJson(Map<String, dynamic> json) => ProductMedia(
        url: json['url']?.toString() ?? '',
        type: json['type']?.toString() ?? 'image',
        alt: json['alt']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'url': url,
        'type': type,
        if (alt != null) 'alt': alt,
      };

  @override
  List<Object?> get props => [url, type, alt];
}

class ProductVariant extends Equatable {
  const ProductVariant({
    required this.id,
    required this.label,
    required this.priceMinor,
    this.compareAtMinor,
    this.currency = 'TRY',
    this.stockStatus = ProductStockStatus.inStock,
    this.sku,
  });

  final String id;
  final String label;
  final int priceMinor;
  final int? compareAtMinor;
  final String currency;
  final ProductStockStatus stockStatus;
  final String? sku;

  factory ProductVariant.fromJson(Map<String, dynamic> json) => ProductVariant(
        id: json['id']?.toString() ?? '',
        label: json['label']?.toString() ?? json['name']?.toString() ?? '',
        priceMinor: (json['price_minor'] as num?)?.toInt() ?? 0,
        compareAtMinor: (json['compare_at_minor'] as num?)?.toInt(),
        currency: json['currency']?.toString() ?? 'TRY',
        stockStatus: productStockStatusFromJson(json['stock_status']?.toString()),
        sku: json['sku']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'label': label,
        'price_minor': priceMinor,
        if (compareAtMinor != null) 'compare_at_minor': compareAtMinor,
        'currency': currency,
        'stock_status': productStockStatusToJson(stockStatus),
        if (sku != null) 'sku': sku,
      };

  @override
  List<Object?> get props => [id, label, priceMinor, compareAtMinor, currency, stockStatus, sku];
}

class ProductBundleItem extends Equatable {
  const ProductBundleItem({required this.productId, required this.quantity, this.title});

  final String productId;
  final int quantity;
  final String? title;

  factory ProductBundleItem.fromJson(Map<String, dynamic> json) => ProductBundleItem(
        productId: json['product_id']?.toString() ?? json['id']?.toString() ?? '',
        quantity: (json['quantity'] as num?)?.toInt() ?? 1,
        title: json['title']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'product_id': productId,
        'quantity': quantity,
        if (title != null) 'title': title,
      };

  @override
  List<Object?> get props => [productId, quantity, title];
}

class ProductBundle extends Equatable {
  const ProductBundle({
    required this.id,
    required this.title,
    required this.items,
    required this.priceMinor,
    this.currency = 'TRY',
  });

  final String id;
  final String title;
  final List<ProductBundleItem> items;
  final int priceMinor;
  final String currency;

  factory ProductBundle.fromJson(Map<String, dynamic> json) => ProductBundle(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => ProductBundleItem.fromJson(e as Map<String, dynamic>))
            .toList(),
        priceMinor: (json['price_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        'items': items.map((e) => e.toJson()).toList(),
        'price_minor': priceMinor,
        'currency': currency,
      };

  @override
  List<Object?> get props => [id, title, items, priceMinor, currency];
}

class ProductNutrition extends Equatable {
  const ProductNutrition({
    this.servingSize,
    this.calories,
    this.proteinGrams,
    this.carbsGrams,
    this.fatGrams,
    this.fiberGrams,
    this.sodiumMg,
    this.facts = const {},
  });

  final String? servingSize;
  final num? calories;
  final num? proteinGrams;
  final num? carbsGrams;
  final num? fatGrams;
  final num? fiberGrams;
  final num? sodiumMg;
  final Map<String, dynamic> facts;

  factory ProductNutrition.fromJson(Map<String, dynamic> json) => ProductNutrition(
        servingSize: json['serving_size']?.toString(),
        calories: json['calories'] as num?,
        proteinGrams: json['protein_grams'] as num?,
        carbsGrams: json['carbs_grams'] as num?,
        fatGrams: json['fat_grams'] as num?,
        fiberGrams: json['fiber_grams'] as num?,
        sodiumMg: json['sodium_mg'] as num?,
        facts: Map<String, dynamic>.from(json['facts'] as Map? ?? {}),
      );

  Map<String, dynamic> toJson() => {
        if (servingSize != null) 'serving_size': servingSize,
        if (calories != null) 'calories': calories,
        if (proteinGrams != null) 'protein_grams': proteinGrams,
        if (carbsGrams != null) 'carbs_grams': carbsGrams,
        if (fatGrams != null) 'fat_grams': fatGrams,
        if (fiberGrams != null) 'fiber_grams': fiberGrams,
        if (sodiumMg != null) 'sodium_mg': sodiumMg,
        if (facts.isNotEmpty) 'facts': facts,
      };

  @override
  List<Object?> get props =>
      [servingSize, calories, proteinGrams, carbsGrams, fatGrams, fiberGrams, sodiumMg, facts];
}

class ProductReviewsSummary extends Equatable {
  const ProductReviewsSummary({
    this.averageRating = 0,
    this.count = 0,
    this.distribution = const {},
  });

  final double averageRating;
  final int count;
  final Map<int, int> distribution;

  factory ProductReviewsSummary.fromJson(Map<String, dynamic> json) {
    final rawDist = json['distribution'] as Map<String, dynamic>? ?? {};
    return ProductReviewsSummary(
      averageRating: (json['average_rating'] as num?)?.toDouble() ?? 0,
      count: (json['count'] as num?)?.toInt() ?? 0,
      distribution: rawDist.map(
        (k, v) => MapEntry(int.tryParse(k) ?? 0, (v as num).toInt()),
      ),
    );
  }

  Map<String, dynamic> toJson() => {
        'average_rating': averageRating,
        'count': count,
        'distribution': distribution.map((k, v) => MapEntry(k.toString(), v)),
      };

  @override
  List<Object?> get props => [averageRating, count, distribution];
}

class ProductQaSummary extends Equatable {
  const ProductQaSummary({this.questionCount = 0, this.answeredCount = 0});

  final int questionCount;
  final int answeredCount;

  factory ProductQaSummary.fromJson(Map<String, dynamic> json) => ProductQaSummary(
        questionCount: (json['question_count'] as num?)?.toInt() ?? 0,
        answeredCount: (json['answered_count'] as num?)?.toInt() ?? 0,
      );

  Map<String, dynamic> toJson() => {
        'question_count': questionCount,
        'answered_count': answeredCount,
      };

  @override
  List<Object?> get props => [questionCount, answeredCount];
}

class ProductQaItem extends Equatable {
  const ProductQaItem({
    required this.question,
    this.answer,
    this.id,
    this.askedAt,
  });

  final String? id;
  final String question;
  final String? answer;
  final DateTime? askedAt;

  factory ProductQaItem.fromJson(Map<String, dynamic> json) => ProductQaItem(
        id: json['id']?.toString(),
        question: json['question']?.toString() ?? '',
        answer: json['answer']?.toString(),
        askedAt: DateTime.tryParse(json['asked_at']?.toString() ?? ''),
      );

  Map<String, dynamic> toJson() => {
        if (id != null) 'id': id,
        'question': question,
        if (answer != null) 'answer': answer,
        if (askedAt != null) 'asked_at': askedAt!.toUtc().toIso8601String(),
      };

  @override
  List<Object?> get props => [id, question, answer, askedAt];
}

class ProductSummary extends Equatable {
  const ProductSummary({
    required this.id,
    required this.title,
    required this.priceMinor,
    this.currency = 'TRY',
    this.imageUrl,
    this.stockStatus = ProductStockStatus.inStock,
  });

  final String id;
  final String title;
  final int priceMinor;
  final String currency;
  final String? imageUrl;
  final ProductStockStatus stockStatus;

  factory ProductSummary.fromJson(Map<String, dynamic> json) => ProductSummary(
        id: json['id']?.toString() ??
            json['productId']?.toString() ??
            json['ProductID']?.toString() ??
            json['sku']?.toString() ??
            json['SKU']?.toString() ??
            '',
        title: json['title']?.toString() ??
            json['Title']?.toString() ??
            json['name']?.toString() ??
            json['sku']?.toString() ??
            '',
        priceMinor: (json['price_minor'] as num?)?.toInt() ??
            (json['priceMinor'] as num?)?.toInt() ??
            0,
        currency: json['currency']?.toString() ?? 'TRY',
        imageUrl: json['image_url']?.toString() ??
            (json['images'] as List<dynamic>?)?.firstOrNull?.toString(),
        stockStatus: productStockStatusFromJson(json['stock_status']?.toString()),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        'price_minor': priceMinor,
        'currency': currency,
        if (imageUrl != null) 'image_url': imageUrl,
        'stock_status': productStockStatusToJson(stockStatus),
      };

  @override
  List<Object?> get props => [id, title, priceMinor, currency, imageUrl, stockStatus];
}

class ProductPricePoint extends Equatable {
  const ProductPricePoint({required this.date, required this.priceMinor, this.currency = 'TRY'});

  final DateTime date;
  final int priceMinor;
  final String currency;

  factory ProductPricePoint.fromJson(Map<String, dynamic> json) => ProductPricePoint(
        date: DateTime.tryParse(json['date']?.toString() ?? '') ?? DateTime.now(),
        priceMinor: (json['price_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
      );

  Map<String, dynamic> toJson() => {
        'date': date.toUtc().toIso8601String(),
        'price_minor': priceMinor,
        'currency': currency,
      };

  @override
  List<Object?> get props => [date, priceMinor, currency];
}

class Product extends Equatable {
  const Product({
    required this.id,
    this.title = '',
    this.description,
    this.images = const [],
    this.videos = const [],
    this.variants = const [],
    this.bundles = const [],
    this.nutrition,
    this.ingredients = const [],
    this.allergens = const [],
    this.origin,
    this.brand,
    this.priceMinor = 0,
    this.compareAtMinor,
    this.currency = 'TRY',
    this.discountPercent,
    this.stockStatus = ProductStockStatus.inStock,
    this.lowStockThreshold,
    this.alternatives = const [],
    this.crossSell = const [],
    this.upsell = const [],
    this.aiRecommendations = const [],
    this.reviewsSummary,
    this.qaSummary,
    this.questions = const [],
    this.isFavorite = false,
    this.dietaryTags = const [],
    this.ageRestricted = false,
    this.weightGrams,
    this.maxQty,
  });

  final String id;
  final String title;
  final String? description;
  final List<String> images;
  final List<ProductMedia> videos;
  final List<ProductVariant> variants;
  final List<ProductBundle> bundles;
  final ProductNutrition? nutrition;
  final List<String> ingredients;
  final List<String> allergens;
  final String? origin;
  final String? brand;
  final int priceMinor;
  final int? compareAtMinor;
  final String currency;
  final int? discountPercent;
  final ProductStockStatus stockStatus;
  final int? lowStockThreshold;
  final List<ProductSummary> alternatives;
  final List<ProductSummary> crossSell;
  final List<ProductSummary> upsell;
  final List<ProductSummary> aiRecommendations;
  final ProductReviewsSummary? reviewsSummary;
  final ProductQaSummary? qaSummary;
  final List<ProductQaItem> questions;
  final bool isFavorite;
  final List<String> dietaryTags;
  final bool ageRestricted;
  final int? weightGrams;
  final int? maxQty;

  String? get primaryImageUrl => images.isNotEmpty ? images.first : null;

  factory Product.fromJson(Map<String, dynamic> json) {
    if (json['product'] is Map) {
      json = Map<String, dynamic>.from(json['product'] as Map);
    }
    List<ProductSummary> parseSummaries(String key) =>
        (json[key] as List<dynamic>? ?? [])
            .map((e) => ProductSummary.fromJson(e as Map<String, dynamic>))
            .toList();

    final imageList = (json['images'] as List<dynamic>? ?? [])
        .map((e) => e is Map ? e['url']?.toString() ?? '' : e.toString())
        .where((e) => e.isNotEmpty)
        .toList();

    return Product(
      id: json['id']?.toString() ??
          json['ID']?.toString() ??
          json['productId']?.toString() ??
          '',
      title: json['title']?.toString() ??
          json['Title']?.toString() ??
          json['name']?.toString() ??
          '',
      description: json['description']?.toString() ??
          json['Description']?.toString(),
      images: imageList,
      videos: (json['videos'] as List<dynamic>? ?? [])
          .map((e) => ProductMedia.fromJson(e as Map<String, dynamic>))
          .toList(),
      variants: (json['variants'] as List<dynamic>? ?? [])
          .map((e) => ProductVariant.fromJson(e as Map<String, dynamic>))
          .toList(),
      bundles: (json['bundles'] as List<dynamic>? ?? [])
          .map((e) => ProductBundle.fromJson(e as Map<String, dynamic>))
          .toList(),
      nutrition: json['nutrition'] != null
          ? ProductNutrition.fromJson(json['nutrition'] as Map<String, dynamic>)
          : null,
      ingredients: (json['ingredients'] as List<dynamic>? ?? [])
          .map((e) => e.toString())
          .toList(),
      allergens: (json['allergens'] as List<dynamic>? ?? [])
          .map((e) => e.toString())
          .toList(),
      origin: json['origin']?.toString(),
      brand: json['brand']?.toString() ?? json['Brand']?.toString(),
      priceMinor: (json['price_minor'] as num?)?.toInt() ??
          (json['priceMinor'] as num?)?.toInt() ??
          0,
      compareAtMinor: (json['compare_at_minor'] as num?)?.toInt(),
      currency: json['currency']?.toString() ?? 'TRY',
      discountPercent: (json['discount_percent'] as num?)?.toInt(),
      stockStatus: productStockStatusFromJson(json['stock_status']?.toString()),
      lowStockThreshold: (json['low_stock_threshold'] as num?)?.toInt(),
      alternatives: parseSummaries('alternatives'),
      crossSell: parseSummaries('cross_sell'),
      upsell: parseSummaries('upsell'),
      aiRecommendations: parseSummaries('ai_recommendations'),
      reviewsSummary: json['reviews_summary'] != null
          ? ProductReviewsSummary.fromJson(json['reviews_summary'] as Map<String, dynamic>)
          : null,
      qaSummary: json['qa_summary'] != null
          ? ProductQaSummary.fromJson(json['qa_summary'] as Map<String, dynamic>)
          : null,
      questions: (json['questions'] as List<dynamic>? ?? [])
          .map((e) => ProductQaItem.fromJson(e as Map<String, dynamic>))
          .toList(),
      isFavorite: json['is_favorite'] == true,
      dietaryTags: (json['dietary_tags'] as List<dynamic>? ?? [])
          .map((e) => e.toString())
          .toList(),
      ageRestricted: json['age_restricted'] == true,
      weightGrams: (json['weight_grams'] as num?)?.toInt(),
      maxQty: (json['max_qty'] as num?)?.toInt(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        if (description != null) 'description': description,
        'images': images,
        'videos': videos.map((e) => e.toJson()).toList(),
        'variants': variants.map((e) => e.toJson()).toList(),
        'bundles': bundles.map((e) => e.toJson()).toList(),
        if (nutrition != null) 'nutrition': nutrition!.toJson(),
        'ingredients': ingredients,
        'allergens': allergens,
        if (origin != null) 'origin': origin,
        if (brand != null) 'brand': brand,
        'price_minor': priceMinor,
        if (compareAtMinor != null) 'compare_at_minor': compareAtMinor,
        'currency': currency,
        if (discountPercent != null) 'discount_percent': discountPercent,
        'stock_status': productStockStatusToJson(stockStatus),
        if (lowStockThreshold != null) 'low_stock_threshold': lowStockThreshold,
        'alternatives': alternatives.map((e) => e.toJson()).toList(),
        'cross_sell': crossSell.map((e) => e.toJson()).toList(),
        'upsell': upsell.map((e) => e.toJson()).toList(),
        'ai_recommendations': aiRecommendations.map((e) => e.toJson()).toList(),
        if (reviewsSummary != null) 'reviews_summary': reviewsSummary!.toJson(),
        if (qaSummary != null) 'qa_summary': qaSummary!.toJson(),
        'questions': questions.map((e) => e.toJson()).toList(),
        'is_favorite': isFavorite,
        'dietary_tags': dietaryTags,
        'age_restricted': ageRestricted,
        if (weightGrams != null) 'weight_grams': weightGrams,
        if (maxQty != null) 'max_qty': maxQty,
      };

  Product copyWith({bool? isFavorite}) => Product(
        id: id,
        title: title,
        description: description,
        images: images,
        videos: videos,
        variants: variants,
        bundles: bundles,
        nutrition: nutrition,
        ingredients: ingredients,
        allergens: allergens,
        origin: origin,
        brand: brand,
        priceMinor: priceMinor,
        compareAtMinor: compareAtMinor,
        currency: currency,
        discountPercent: discountPercent,
        stockStatus: stockStatus,
        lowStockThreshold: lowStockThreshold,
        alternatives: alternatives,
        crossSell: crossSell,
        upsell: upsell,
        aiRecommendations: aiRecommendations,
        reviewsSummary: reviewsSummary,
        qaSummary: qaSummary,
        questions: questions,
        isFavorite: isFavorite ?? this.isFavorite,
        dietaryTags: dietaryTags,
        ageRestricted: ageRestricted,
        weightGrams: weightGrams,
        maxQty: maxQty,
      );

  @override
  List<Object?> get props => [
        id,
        title,
        description,
        images,
        videos,
        variants,
        bundles,
        nutrition,
        ingredients,
        allergens,
        origin,
        brand,
        priceMinor,
        compareAtMinor,
        currency,
        discountPercent,
        stockStatus,
        lowStockThreshold,
        alternatives,
        crossSell,
        upsell,
        aiRecommendations,
        reviewsSummary,
        qaSummary,
        questions,
        isFavorite,
        dietaryTags,
        ageRestricted,
        weightGrams,
        maxQty,
      ];
}
