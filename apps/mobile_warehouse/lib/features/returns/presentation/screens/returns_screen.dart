import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/return_task.dart';
import '../providers/returns_providers.dart';

class ReturnsScreen extends ConsumerWidget {
  const ReturnsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(returnsListProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Returns'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(returnsListProvider),
        child: AsyncValueWidget<List<ReturnTask>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [SizedBox(height: 120), NxEmptyState(title: 'No returns')]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final t = items[i];
                return NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('${t.type.toUpperCase()} · ${t.id}', style: NxTypography.titleSm),
                      Text('${t.reference ?? '-'} · ${t.sku ?? '-'} × ${t.qty} · ${t.status}',
                          style: NxTypography.captionMd),
                      const SizedBox(height: NxSpacing.s2),
                      Wrap(
                        spacing: NxSpacing.s2,
                        children: [
                          NxButton(label: 'Inspect', size: NxButtonSize.sm, onPressed: () => ref.read(returnsActionsProvider).advance(t.id, 'inspect')),
                          NxButton(label: 'Restock', size: NxButtonSize.sm, variant: NxButtonVariant.secondary, onPressed: () => ref.read(returnsActionsProvider).advance(t.id, 'restock')),
                          NxButton(label: 'Dispose', size: NxButtonSize.sm, variant: NxButtonVariant.destructive, onPressed: () => ref.read(returnsActionsProvider).advance(t.id, 'dispose')),
                        ],
                      ),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}
