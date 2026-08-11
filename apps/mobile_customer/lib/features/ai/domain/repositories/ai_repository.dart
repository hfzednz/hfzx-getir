import 'package:nexora_core/nexora_core.dart';

import '../entities/ai_entity.dart';

abstract class AiRepository {
  Future<Result<AiRecommendation>> refreshRecommendations({Map<String, dynamic>? context});
  Future<Result<List<AiRecipeSuggestion>>> recipeSuggestions({String? query});
  Future<Result<List<String>>> reorderPrediction();
  Future<Result<AiBudgetSuggestion>> budgetOptimization({required int budgetMinor});
  Future<Result<List<String>>> nutritionSuggestions({List<String>? productIds});
  Future<Result<AiAssistantReply>> shoppingAssistantMessage({
    required String message,
    Map<String, dynamic>? context,
  });
}
