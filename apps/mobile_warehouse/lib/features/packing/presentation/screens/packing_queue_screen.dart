import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/packing_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/pack_task.dart';
import '../providers/packing_providers.dart';

class PackingQueueScreen extends ConsumerWidget {
  const PackingQueueScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(packingQueueProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Packing'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(packingQueueProvider),
        child: AsyncValueWidget<List<PackTask>>(
          value: async,
          data: (tasks) {
            if (tasks.isEmpty) {
              return ListView(children: const [
                SizedBox(height: 120),
                NxEmptyState(title: 'Queue empty', body: 'No pack tasks.'),
              ]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: tasks.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final t = tasks[i];
                return NxCard(
                  variant: NxCardVariant.interactive,
                  onTap: () => context.push(RouteNames.packingTaskPath(t.id)),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Order ${t.orderId}', style: NxTypography.titleSm),
                            Text('${t.itemCount} items · ${t.status.wireName}',
                                style: NxTypography.captionMd),
                          ],
                        ),
                      ),
                      if (PackingRules.canClaim(t.status))
                        NxButton(
                          label: 'Claim',
                          size: NxButtonSize.sm,
                          onPressed: () => ref.read(packingActionsProvider).claim(t.id),
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
