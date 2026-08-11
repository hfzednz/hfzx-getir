import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/purchasing_entities.dart';
import '../providers/purchasing_providers.dart';

class PurchasingScreen extends ConsumerWidget {
  const PurchasingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final suppliers = ref.watch(suppliersProvider);
    final orders = ref.watch(purchaseOrdersProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Purchasing'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s3),
        children: [
          Text('Suppliers', style: NxTypography.titleSm),
          const SizedBox(height: NxSpacing.s2),
          AsyncValueWidget<List<Supplier>>(
            value: suppliers,
            data: (items) => Column(
              children: items.map((s) => NxCard(
                child: Text('${s.name}${s.contact != null ? ' · ${s.contact}' : ''}', style: NxTypography.bodySm),
              )).toList(),
            ),
          ),
          const SizedBox(height: NxSpacing.s4),
          Text('Purchase orders', style: NxTypography.titleSm),
          const SizedBox(height: NxSpacing.s2),
          AsyncValueWidget<List<PurchaseOrder>>(
            value: orders,
            data: (items) {
              if (items.isEmpty) return const NxEmptyState(title: 'No POs');
              return Column(
                children: items.map((po) => Padding(
                  padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                  child: NxCard(
                    child: Row(
                      children: [
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(po.id, style: NxTypography.titleSm),
                              Text('${po.supplierName ?? po.supplierId} · ${po.status} · ${po.lineCount} lines',
                                  style: NxTypography.captionMd),
                            ],
                          ),
                        ),
                        NxButton(
                          label: 'Receive+QC',
                          size: NxButtonSize.sm,
                          onPressed: () => ref.read(purchasingActionsProvider).receivePo(po.id, qcFlag: true),
                        ),
                      ],
                    ),
                  ),
                )).toList(),
              );
            },
          ),
        ],
      ),
    );
  }
}
