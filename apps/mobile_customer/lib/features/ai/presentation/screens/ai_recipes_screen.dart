import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../domain/entities/ai_entity.dart';
import '../providers/ai_providers.dart';

class AiRecipesScreen extends ConsumerWidget {
  const AiRecipesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final recipesAsync = ref.watch(aiRecipeSuggestionsProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'AI Recipes'),
      body: recipesAsync.when(
        data: (recipes) {
          if (recipes.isEmpty) {
            return const NxEmptyState(
              title: 'No recipes',
              body: 'Try again later for meal suggestions.',
            );
          }
          return ListView.separated(
            padding: const EdgeInsets.all(NxSpacing.s4),
            itemCount: recipes.length,
            separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
            itemBuilder: (context, index) {
              final recipe = recipes[index];
              return _RecipeCard(recipe: recipe);
            },
          );
        },
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}

class _RecipeCard extends ConsumerStatefulWidget {
  const _RecipeCard({required this.recipe});

  final AiRecipeSuggestion recipe;

  @override
  ConsumerState<_RecipeCard> createState() => _RecipeCardState();
}

class _RecipeCardState extends ConsumerState<_RecipeCard> {
  bool _busy = false;

  Future<void> _addToCart() async {
    setState(() => _busy = true);
    final message = await addRecipeIngredientsToCart(ref, widget.recipe);
    if (!mounted) return;
    setState(() => _busy = false);
    if (message != null) {
      NxToast.show(context, message: message, variant: NxToastVariant.warning);
    } else {
      NxToast.show(
        context,
        message: 'Ingredients added to cart',
        variant: NxToastVariant.success,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final recipe = widget.recipe;
    return NxCard(
      child: Padding(
        padding: const EdgeInsets.all(NxSpacing.s3),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(recipe.title, style: NxTypography.headlineSm),
            if (recipe.instructions.isNotEmpty) ...[
              const SizedBox(height: NxSpacing.s2),
              Text(recipe.instructions, style: NxTypography.bodyMd),
            ],
            if (recipe.items.isNotEmpty) ...[
              const SizedBox(height: NxSpacing.s3),
              Text('Ingredients', style: NxTypography.captionMd),
              ...recipe.items.map((item) {
                final priceLabel = item.priceMinor != null && item.priceMinor! > 0
                    ? Formatters.money(
                        Money(minorUnits: item.priceMinor!, currency: item.currency),
                      )
                    : 'Price unavailable';
                return ListTile(
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  title: Text(item.title ?? item.productId),
                  subtitle: Text('×${item.quantity}'),
                  trailing: Text(priceLabel),
                );
              }),
            ] else if (recipe.ingredientProductIds.isNotEmpty) ...[
              const SizedBox(height: NxSpacing.s2),
              Text(
                '${recipe.ingredientProductIds.length} ingredients (prices unavailable)',
                style: NxTypography.captionMd,
              ),
            ],
            const SizedBox(height: NxSpacing.s3),
            NxButton(
              label: 'Add ingredients to cart',
              expand: true,
              loading: _busy,
              onPressed: _addToCart,
            ),
          ],
        ),
      ),
    );
  }
}
