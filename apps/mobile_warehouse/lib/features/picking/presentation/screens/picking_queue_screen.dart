import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/picking_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/pick_task.dart';
import '../providers/picking_providers.dart';

class PickingQueueScreen extends ConsumerWidget {
  const PickingQueueScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(pickingQueueProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Picking'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(pickingQueueProvider),
        child: AsyncValueWidget<List<PickTask>>(
          value: async,
          data: (tasks) {
            if (tasks.isEmpty) {
              return ListView(
                children: const [
                  SizedBox(height: 120),
                  NxEmptyState(
                    title: 'Queue empty',
                    body: 'No pick tasks right now.',
                  ),
                ],
              );
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: tasks.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final t = tasks[i];
                return NxCard(
                  variant: NxCardVariant.interactive,
                  onTap: () => context.push(RouteNames.pickingTaskPath(t.id)),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Order ${t.orderId}',
                              style: NxTypography.titleSm,
                            ),
                            Text(
                              '${t.lines.length} lines · ${t.status.wireName}'
                              '${t.zone != null ? ' · ${t.zone}' : ''}',
                              style: NxTypography.captionMd,
                            ),
                          ],
                        ),
                      ),
                      if (PickingRules.canClaim(t.status))
                        NxButton(
                          label: 'Claim',
                          size: NxButtonSize.sm,
                          onPressed: () async {
                            await ref.read(pickingActionsProvider).claim(t.id);
                          },
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
