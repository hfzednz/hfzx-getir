import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/business_rules/packing_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/pack_task.dart';
import '../providers/packing_providers.dart';

class PackTaskScreen extends ConsumerStatefulWidget {
  const PackTaskScreen({super.key, required this.taskId});
  final String taskId;

  @override
  ConsumerState<PackTaskScreen> createState() => _PackTaskScreenState();
}

class _PackTaskScreenState extends ConsumerState<PackTaskScreen> {
  final _weightCtrl = TextEditingController();

  @override
  void dispose() {
    _weightCtrl.dispose();
    super.dispose();
  }

  Future<void> _printIntent(PackTask task) async {
    final result = await ref.read(packingActionsProvider).printLabel(task.id);
    if (!mounted) return;
    result.fold(
      onSuccess: (updated) {
        final url = updated.labelUrl ?? task.labelUrl;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              url != null
                  ? 'Print intent: $url'
                  : 'Label print requested',
            ),
          ),
        );
      },
      onFailure: (e) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message)),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(packTaskProvider(widget.taskId));
    final actions = ref.read(packingActionsProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Pack station'),
      body: AsyncValueWidget<PackTask>(
        value: async,
        data: (task) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s3),
          children: [
            NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Order ${task.orderId}', style: NxTypography.titleMd),
                  Text('Status: ${task.status.wireName}', style: NxTypography.captionMd),
                  Text(
                    'Expected weight: ${task.expectedWeightGrams.toStringAsFixed(0)} g',
                    style: NxTypography.bodySm,
                  ),
                ],
              ),
            ),
            const SizedBox(height: NxSpacing.s3),
            Text('Materials', style: NxTypography.titleSm),
            const SizedBox(height: NxSpacing.s2),
            ...task.materials.map(
              (m) => Padding(
                padding: const EdgeInsets.only(bottom: NxSpacing.s1),
                child: Text('${m.qty}× ${m.name} (${m.code})', style: NxTypography.bodySm),
              ),
            ),
            if (PackingRules.canWeigh(task.status)) ...[
              const SizedBox(height: NxSpacing.s3),
              TextField(
                controller: _weightCtrl,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Actual weight (g)'),
              ),
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Weight check',
                expand: true,
                onPressed: () {
                  final g = double.tryParse(_weightCtrl.text) ?? 0;
                  actions.weigh(task, g);
                },
              ),
            ],
            if (PackingRules.canPrintLabel(task.status)) ...[
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Print label',
                expand: true,
                variant: NxButtonVariant.secondary,
                onPressed: () => _printIntent(task),
              ),
            ],
            if (PackingRules.canSeal(task.status) || task.status == PackTaskStatus.labeled) ...[
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Confirm seal',
                expand: true,
                onPressed: () => actions.seal(task),
              ),
            ],
            if (PackingRules.canComplete(task.status)) ...[
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Send to dispatch',
                expand: true,
                onPressed: () => actions.complete(task.id),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
