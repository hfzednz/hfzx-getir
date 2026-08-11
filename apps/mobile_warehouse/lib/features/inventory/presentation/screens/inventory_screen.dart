import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/inventory_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/stock_item.dart';
import '../providers/inventory_providers.dart';

class InventoryScreen extends ConsumerWidget {
  const InventoryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(stockListProvider);
    return Scaffold(
      appBar: NxTopBar(
        title: 'Inventory',
        actions: [
          IconButton(
            icon: const Icon(Icons.fact_check_outlined),
            onPressed: () => context.push(RouteNames.cycleCount),
          ),
          IconButton(
            icon: const Icon(Icons.move_to_inbox_outlined),
            onPressed: () => context.push(RouteNames.inboundReceive),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(stockListProvider),
        child: AsyncValueWidget<List<StockItem>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [
                SizedBox(height: 120),
                NxEmptyState(title: 'No stock', body: 'Stock list is empty.'),
              ]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final item = items[i];
                final low = InventoryRules.isLowStock(
                  onHand: item.onHand,
                  reorderPoint: item.reorderPoint,
                );
                final oos = InventoryRules.isOutOfStock(item.onHand);
                return NxCard(
                  variant: NxCardVariant.interactive,
                  onTap: () => context.push(RouteNames.inventoryAdjustPath(item.sku)),
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(item.name, style: NxTypography.titleSm),
                            Text(
                              '${item.sku} · ${item.bin ?? '-'} · on hand ${item.onHand}',
                              style: NxTypography.captionMd,
                            ),
                          ],
                        ),
                      ),
                      if (oos)
                        Text('OOS', style: NxTypography.captionMd.copyWith(color: context.nxColors.danger))
                      else if (low)
                        Text('LOW', style: NxTypography.captionMd.copyWith(color: context.nxColors.warning)),
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
