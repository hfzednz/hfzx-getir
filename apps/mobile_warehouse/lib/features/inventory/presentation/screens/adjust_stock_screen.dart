import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/stock_item.dart';
import '../providers/inventory_providers.dart';

class AdjustStockScreen extends ConsumerStatefulWidget {
  const AdjustStockScreen({super.key, required this.sku});
  final String sku;

  @override
  ConsumerState<AdjustStockScreen> createState() => _AdjustStockScreenState();
}

class _AdjustStockScreenState extends ConsumerState<AdjustStockScreen> {
  final _deltaCtrl = TextEditingController();
  String _reason = 'cycle_count';

  @override
  void dispose() {
    _deltaCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(stockListProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Adjust stock'),
      body: AsyncValueWidget<List<StockItem>>(
        value: async,
        data: (items) {
          StockItem? item;
          for (final e in items) {
            if (e.sku == widget.sku) {
              item = e;
              break;
            }
          }
          if (item == null) {
            return const NxEmptyState(title: 'SKU not found');
          }
          final stock = item;
          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s3),
            children: [
              Text('${stock.name} (${stock.sku})', style: NxTypography.titleMd),
              Text('On hand: ${stock.onHand}', style: NxTypography.bodySm),
              const SizedBox(height: NxSpacing.s3),
              TextField(
                controller: _deltaCtrl,
                keyboardType: TextInputType.number,
                decoration: const InputDecoration(labelText: 'Delta (+/-)'),
              ),
              const SizedBox(height: NxSpacing.s2),
              DropdownButtonFormField<String>(
                initialValue: _reason,
                items: const [
                  DropdownMenuItem(value: 'damage', child: Text('Damage')),
                  DropdownMenuItem(value: 'expiry', child: Text('Expiry')),
                  DropdownMenuItem(value: 'found', child: Text('Found')),
                  DropdownMenuItem(value: 'cycle_count', child: Text('Cycle count')),
                  DropdownMenuItem(value: 'other', child: Text('Other')),
                ],
                onChanged: (v) => setState(() => _reason = v ?? _reason),
                decoration: const InputDecoration(labelText: 'Reason'),
              ),
              const SizedBox(height: NxSpacing.s3),
              NxButton(
                label: 'Submit adjust',
                expand: true,
                onPressed: () async {
                  final delta = int.tryParse(_deltaCtrl.text) ?? 0;
                  final r = await ref.read(inventoryActionsProvider).adjust(
                        item: stock,
                        delta: delta,
                        reasonCode: _reason,
                      );
                  if (!context.mounted) return;
                  r.fold(
                    onSuccess: (_) => context.pop(),
                    onFailure: (e) => ScaffoldMessenger.of(context)
                        .showSnackBar(SnackBar(content: Text(e.message))),
                  );
                },
              ),
            ],
          );
        },
      ),
    );
  }
}
