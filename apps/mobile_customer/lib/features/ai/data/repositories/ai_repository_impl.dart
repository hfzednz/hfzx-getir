import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/ai_entity.dart';
import '../../domain/repositories/ai_repository.dart';
import '../../data/datasources/ai_remote_datasource.dart';

class AiRepositoryImpl implements AiRepository {
  const AiRepositoryImpl(this._remote);
  final AiRemoteDataSource _remote;

  @override
  Future<Result<AiRecommendation>> refreshRecommendations({Map<String, dynamic>? context}) =>
      _remote.refreshRecommendations(context: context);

  @override
  Future<Result<List<AiRecipeSuggestion>>> recipeSuggestions({String? query}) =>
      _remote.recipeSuggestions(query: query);

  @override
  Future<Result<List<String>>> reorderPrediction() => _remote.reorderPrediction();

  @override
  Future<Result<AiBudgetSuggestion>> budgetOptimization({required int budgetMinor}) =>
      _remote.budgetOptimization(budgetMinor: budgetMinor);

  @override
  Future<Result<List<String>>> nutritionSuggestions({List<String>? productIds}) =>
      _remote.nutritionSuggestions(productIds: productIds);

  @override
  Future<Result<AiAssistantReply>> shoppingAssistantMessage({
    required String message,
    Map<String, dynamic>? context,
  }) =>
      _remote.shoppingAssistantMessage(message: message, context: context);
}
