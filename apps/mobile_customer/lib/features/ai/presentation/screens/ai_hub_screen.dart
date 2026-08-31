import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../providers/ai_providers.dart';
import '../../../../shared/errors/error_copy.dart';

class AiHubScreen extends ConsumerStatefulWidget {
  const AiHubScreen({super.key});

  @override
  ConsumerState<AiHubScreen> createState() => _AiHubScreenState();
}

class _AiHubScreenState extends ConsumerState<AiHubScreen> {
  final _assistantController = TextEditingController();

  @override
  void dispose() {
    _assistantController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final recommendations = ref.watch(aiRecommendationsProvider);
    final reorder = ref.watch(aiReorderPredictionProvider);
    final assistant = ref.watch(aiAssistantReplyProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'AI Hub'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text('Recommendations', style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s2),
          recommendations.when(
            data: (rec) => Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (rec.reason.isNotEmpty)
                  Text(rec.reason, style: NxTypography.bodyMd),
                const SizedBox(height: NxSpacing.s2),
                if (rec.productIds.isEmpty)
                  Text(
                    'No recommendations yet',
                    style: NxTypography.captionMd,
                  )
                else
                  Wrap(
                    spacing: NxSpacing.s2,
                    children: rec.productIds
                        .map(
                          (id) => NxChip(
                            label: id,
                            onSelected: (_) => context.push('/p/$id'),
                          ),
                        )
                        .toList(),
                  ),
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: 'Refresh',
                  variant: NxButtonVariant.secondary,
                  onPressed: () => ref.invalidate(aiRecommendationsProvider),
                ),
              ],
            ),
            loading: () => const Center(child: NxSpinner()),
            error: (e, _) => Text(localizedCustomerError(context, e)),
          ),
          const SizedBox(height: NxSpacing.s6),
          Text('Recipes', style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s2),
          Text(
            'Discover meal ideas and add ingredients to your cart.',
            style: NxTypography.bodyMd,
          ),
          const SizedBox(height: NxSpacing.s2),
          NxButton(
            label: 'Browse recipes',
            expand: true,
            onPressed: () => context.push(RouteNames.aiRecipes),
          ),
          const SizedBox(height: NxSpacing.s6),
          Text('Shopping assistant', style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s2),
          NxField(
            label: 'Ask anything',
            controller: _assistantController,
          ),
          const SizedBox(height: NxSpacing.s2),
          NxButton(
            label: 'Ask',
            variant: NxButtonVariant.secondary,
            loading: assistant.isLoading,
            onPressed: () {
              final msg = _assistantController.text.trim();
              if (msg.isEmpty) return;
              ref.read(aiAssistantReplyProvider.notifier).ask(msg);
            },
          ),
          const SizedBox(height: NxSpacing.s2),
          assistant.when(
            data: (reply) {
              if (reply == null) return const SizedBox.shrink();
              return Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(reply.message, style: NxTypography.bodyMd),
                  if (reply.suggestedProductIds.isNotEmpty) ...[
                    const SizedBox(height: NxSpacing.s2),
                    Wrap(
                      spacing: NxSpacing.s2,
                      children: reply.suggestedProductIds
                          .map(
                            (id) => NxChip(
                              label: id,
                              onSelected: (_) => context.push('/p/$id'),
                            ),
                          )
                          .toList(),
                    ),
                  ],
                ],
              );
            },
            loading: () => const Center(child: NxSpinner()),
            error: (e, _) => Text(localizedCustomerError(context, e)),
          ),
          const SizedBox(height: NxSpacing.s6),
          Text('Reorder prediction', style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s2),
          reorder.when(
            data: (ids) {
              if (ids.isEmpty) {
                return Text(
                  'No reorder suggestions',
                  style: NxTypography.captionMd,
                );
              }
              return Wrap(
                spacing: NxSpacing.s2,
                children: ids
                    .map(
                      (id) => NxChip(
                        label: id,
                        onSelected: (_) => context.push('/p/$id'),
                      ),
                    )
                    .toList(),
              );
            },
            loading: () => const Center(child: NxSpinner()),
            error: (e, _) => Text(localizedCustomerError(context, e)),
          ),
          const SizedBox(height: NxSpacing.s2),
          NxButton(
            label: 'Refresh predictions',
            variant: NxButtonVariant.tertiary,
            onPressed: () => ref.invalidate(aiReorderPredictionProvider),
          ),
        ],
      ),
    );
  }
}
