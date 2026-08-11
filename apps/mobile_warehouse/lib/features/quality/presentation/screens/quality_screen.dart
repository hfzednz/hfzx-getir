import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/qc_inspection.dart';
import '../providers/quality_providers.dart';

class QualityScreen extends ConsumerWidget {
  const QualityScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(qualityQueueProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Quality'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(qualityQueueProvider),
        child: AsyncValueWidget<List<QcInspection>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [SizedBox(height: 120), NxEmptyState(title: 'QC queue empty')]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final q = items[i];
                return NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('${q.stage} · ${q.reference ?? q.id}', style: NxTypography.titleSm),
                      Text(q.status, style: NxTypography.captionMd),
                      const SizedBox(height: NxSpacing.s2),
                      Row(
                        children: [
                          Expanded(child: NxButton(label: 'Pass', size: NxButtonSize.sm, onPressed: () => ref.read(qualityActionsProvider).decide(q.id, pass: true))),
                          const SizedBox(width: NxSpacing.s2),
                          Expanded(child: NxButton(label: 'Fail', size: NxButtonSize.sm, variant: NxButtonVariant.destructive, onPressed: () => ref.read(qualityActionsProvider).decide(q.id, pass: false))),
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
