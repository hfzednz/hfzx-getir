import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../cart/presentation/providers/cart_providers.dart';
import '../../data/datasources/ai_remote_datasource.dart';
import '../../data/repositories/ai_repository_impl.dart';
import '../../domain/entities/ai_entity.dart';
import '../../domain/repositories/ai_repository.dart';

final aiRemoteDataSourceProvider = Provider<AiRemoteDataSource>((ref) {
  return AiRemoteDataSource(ref.watch(apiClientProvider));
});

final aiRepositoryProvider = Provider<AiRepository>((ref) {
  return AiRepositoryImpl(ref.watch(aiRemoteDataSourceProvider));
});

final aiRecommendationsProvider = FutureProvider.autoDispose<AiRecommendation>((ref) async {
  final result = await ref.watch(aiRepositoryProvider).refreshRecommendations();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final aiRecipeSuggestionsProvider = FutureProvider.autoDispose<List<AiRecipeSuggestion>>((ref) async {
  final result = await ref.watch(aiRepositoryProvider).recipeSuggestions();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final aiReorderPredictionProvider = FutureProvider.autoDispose<List<String>>((ref) async {
  final result = await ref.watch(aiRepositoryProvider).reorderPrediction();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

Future<String?> addRecipeIngredientsToCart(WidgetRef ref, AiRecipeSuggestion recipe) async {
  final cart = ref.read(cartRepositoryProvider);
  var added = 0;
  var skipped = 0;

  if (recipe.items.isNotEmpty) {
    for (final item in recipe.items) {
      final price = item.priceMinor ?? 0;
      if (price <= 0 || item.productId.isEmpty) {
        skipped++;
        continue;
      }
      await cart.addItem(
        productId: item.productId,
        title: item.title ?? item.productId,
        imageUrl: item.imageUrl,
        unitPriceMinor: price,
        currency: item.currency,
        quantity: item.quantity,
      );
      added++;
    }
  } else {
    // No priced line items — skip zero-price placeholders.
    skipped = recipe.ingredientProductIds.length;
  }

  if (skipped > 0 && added == 0) {
    return 'Could not add ingredients — prices unavailable.';
  }
  if (skipped > 0) {
    return 'Added $added ingredients. Skipped $skipped without prices.';
  }
  return null;
}

final aiAssistantReplyProvider =
    AsyncNotifierProvider<AiAssistantController, AiAssistantReply?>(AiAssistantController.new);

class AiAssistantController extends AsyncNotifier<AiAssistantReply?> {
  @override
  Future<AiAssistantReply?> build() async => null;

  Future<void> ask(String message, {Map<String, dynamic>? context}) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final result = await ref.read(aiRepositoryProvider).shoppingAssistantMessage(
            message: message,
            context: context,
          );
      return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
    });
  }
}
