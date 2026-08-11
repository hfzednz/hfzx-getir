import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/picking_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/scan_fab.dart';
import '../../domain/entities/pick_task.dart';
import '../providers/picking_providers.dart';

class PickTaskScreen extends ConsumerWidget {
  const PickTaskScreen({super.key, required this.taskId});

  final String taskId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(pickTaskProvider(taskId));
    final actions = ref.read(pickingActionsProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Pick task'),
      floatingActionButton: async.maybeWhen(
        data: (task) => PickingRules.canScanLine(task.status)
            ? ScanFab(
                onPressed: () =>
                    context.push(RouteNames.pickingScanPath(taskId)),
              )
            : null,
        orElse: () => null,
      ),
      body: AsyncValueWidget<PickTask>(
        value: async,
        data: (task) {
          final lines = task.pathOrderedLines;
          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.all(NxSpacing.s3),
                child: NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Order ${task.orderId}', style: NxTypography.titleMd),
                      Text(
                        'Status: ${task.status.wireName}',
                        style: NxTypography.captionMd,
                      ),
                      if (task.pathHint != null)
                        Text(task.pathHint!, style: NxTypography.bodySm),
                    ],
                  ),
                ),
              ),
              Expanded(
                child: ListView.separated(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s3),
                  itemCount: lines.length,
                  separatorBuilder: (_, __) =>
                      const SizedBox(height: NxSpacing.s2),
                  itemBuilder: (context, i) {
                    final line = lines[i];
                    return NxCard(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            '${line.pathStep ?? i + 1}. ${line.name ?? line.sku}',
                            style: NxTypography.titleSm,
                          ),
                          Text(
                            'Bin ${line.bin} · ${line.zone ?? '-'} · '
                            '${line.pickedQty}/${line.qty}',
                            style: NxTypography.captionMd,
                          ),
                          if (!line.isComplete &&
                              PickingRules.canShortPick(task.status))
                            Align(
                              alignment: Alignment.centerRight,
                              child: NxButton(
                                label: 'Short-pick',
                                size: NxButtonSize.sm,
                                variant: NxButtonVariant.secondary,
                                onPressed: () async {
                                  final missing = line.qty - line.pickedQty;
                                  await actions.shortPick(
                                    task: task,
                                    line: line,
                                    missingQty: missing,
                                  );
                                },
                              ),
                            ),
                        ],
                      ),
                    );
                  },
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(NxSpacing.s3),
                child: Column(
                  children: [
                    if (PickingRules.canStart(task.status))
                      NxButton(
                        label: 'Start picking',
                        expand: true,
                        onPressed: () => actions.start(taskId),
                      ),
                    if (PickingRules.canComplete(task)) ...[
                      const SizedBox(height: NxSpacing.s2),
                      NxButton(
                        label: 'Complete',
                        expand: true,
                        onPressed: () => actions.complete(taskId),
                      ),
                    ],
                    if (PickingRules.canStage(task.status)) ...[
                      const SizedBox(height: NxSpacing.s2),
                      NxButton(
                        label: 'Stage to pack',
                        expand: true,
                        onPressed: () => actions.stage(taskId),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}
