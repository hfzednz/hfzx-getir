import 'package:equatable/equatable.dart';

class AiRecommendation extends Equatable {
  const AiRecommendation({required this.productIds, this.reason = ''});
  final List<String> productIds;
  final String reason;
  factory AiRecommendation.fromJson(Map<String, dynamic> json) => AiRecommendation(
        productIds: (json['product_ids'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? const [],
        reason: json['reason']?.toString() ?? '',
      );
  @override
  List<Object?> get props => [productIds, reason];
}

class AiRecipeIngredient extends Equatable {
  const AiRecipeIngredient({
    required this.productId,
    this.title,
    this.priceMinor,
    this.currency = 'TRY',
    this.quantity = 1,
    this.imageUrl,
  });

  final String productId;
  final String? title;
  final int? priceMinor;
  final String currency;
  final int quantity;
  final String? imageUrl;

  factory AiRecipeIngredient.fromJson(Map<String, dynamic> json) => AiRecipeIngredient(
        productId: json['product_id']?.toString() ?? json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString(),
        priceMinor: (json['price_minor'] as num?)?.toInt() ??
            (json['unit_price_minor'] as num?)?.toInt(),
        currency: json['currency']?.toString() ?? 'TRY',
        quantity: (json['quantity'] as num?)?.toInt() ?? 1,
        imageUrl: json['image_url']?.toString(),
      );

  @override
  List<Object?> get props => [productId, title, priceMinor, currency, quantity, imageUrl];
}

class AiRecipeSuggestion extends Equatable {
  const AiRecipeSuggestion({
    required this.title,
    required this.ingredientProductIds,
    this.instructions = '',
    this.items = const [],
  });

  final String title;
  final List<String> ingredientProductIds;
  final String instructions;
  final List<AiRecipeIngredient> items;

  factory AiRecipeSuggestion.fromJson(Map<String, dynamic> json) {
    final items = (json['items'] as List<dynamic>? ?? json['ingredients'] as List<dynamic>? ?? [])
        .map((e) => AiRecipeIngredient.fromJson(e as Map<String, dynamic>))
        .toList();
    final ids = (json['ingredient_product_ids'] as List<dynamic>?)
            ?.map((e) => e.toString())
            .toList() ??
        items.map((e) => e.productId).where((id) => id.isNotEmpty).toList();
    return AiRecipeSuggestion(
      title: json['title']?.toString() ?? '',
      ingredientProductIds: ids,
      instructions: json['instructions']?.toString() ?? '',
      items: items,
    );
  }

  @override
  List<Object?> get props => [title, ingredientProductIds, items];
}

class AiBudgetSuggestion extends Equatable {
  const AiBudgetSuggestion({required this.suggestedProductIds, required this.estimatedSavingsMinor});
  final List<String> suggestedProductIds;
  final int estimatedSavingsMinor;
  factory AiBudgetSuggestion.fromJson(Map<String, dynamic> json) => AiBudgetSuggestion(
        suggestedProductIds:
            (json['suggested_product_ids'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? const [],
        estimatedSavingsMinor: (json['estimated_savings_minor'] as num?)?.toInt() ?? 0,
      );
  @override
  List<Object?> get props => [suggestedProductIds, estimatedSavingsMinor];
}

class AiAssistantReply extends Equatable {
  const AiAssistantReply({required this.message, this.suggestedProductIds = const []});
  final String message;
  final List<String> suggestedProductIds;
  factory AiAssistantReply.fromJson(Map<String, dynamic> json) => AiAssistantReply(
        message: json['message']?.toString() ?? json['content']?.toString() ?? '',
        suggestedProductIds:
            (json['suggested_product_ids'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? const [],
      );
  @override
  List<Object?> get props => [message, suggestedProductIds];
}
