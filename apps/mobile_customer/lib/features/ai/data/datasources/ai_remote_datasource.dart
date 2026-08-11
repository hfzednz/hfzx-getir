import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/ai_entity.dart';

class AiRemoteDataSource {
  const AiRemoteDataSource(this._client);
  final ApiClient _client;

  static const _base = '/v1/ai';

  Future<Result<AiRecommendation>> refreshRecommendations({Map<String, dynamic>? context}) async {
    return _client.post<AiRecommendation>(
      '$_base/recommendations/refresh',
      data: context ?? const {},
      parser: (json) => AiRecommendation.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Result<List<AiRecipeSuggestion>>> recipeSuggestions({String? query}) async {
    return _client.post<List<AiRecipeSuggestion>>(
      '$_base/recipes/suggest',
      data: {if (query != null) 'query': query},
      parser: (json) => (json as List<dynamic>)
          .map((e) => AiRecipeSuggestion.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<Result<List<String>>> reorderPrediction() async {
    return _client.get<List<String>>(
      '$_base/reorder/prediction',
      parser: (json) => (json as List<dynamic>).map((e) => e.toString()).toList(),
    );
  }

  Future<Result<AiBudgetSuggestion>> budgetOptimization({required int budgetMinor}) async {
    return _client.post<AiBudgetSuggestion>(
      '$_base/budget/optimize',
      data: {'budget_minor': budgetMinor},
      parser: (json) => AiBudgetSuggestion.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Result<List<String>>> nutritionSuggestions({List<String>? productIds}) async {
    return _client.post<List<String>>(
      '$_base/nutrition/suggest',
      data: {if (productIds != null) 'product_ids': productIds},
      parser: (json) => (json as List<dynamic>).map((e) => e.toString()).toList(),
    );
  }

  Future<Result<AiAssistantReply>> shoppingAssistantMessage({
    required String message,
    Map<String, dynamic>? context,
  }) async {
    return _client.post<AiAssistantReply>(
      '$_base/assistant/message',
      data: {'message': message, ...?context},
      parser: (json) => AiAssistantReply.fromJson(json as Map<String, dynamic>),
    );
  }
}
