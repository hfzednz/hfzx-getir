import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/expiry_item.dart';
import '../providers/expiry_providers.dart';

class ExpiryScreen extends ConsumerWidget {
  const ExpiryScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(nearExpiryProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Near expiry'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(nearExpiryProvider),
        child: AsyncValueWidget<List<ExpiryItem>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [SizedBox(height: 120), NxEmptyState(title: 'No near-expiry items')]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final item = items[i];
                return NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(item.name, style: NxTypography.titleSm),
                      Text('${item.sku} · ${item.bin ?? '-'} · qty ${item.qty}', style: NxTypography.captionMd),
                      Text('Expires ${item.expiresAt.toIso8601String().split('T').first}', style: NxTypography.bodySm),
                      if (item.fefoHint != null)
                        Text('FEFO: ${item.fefoHint}', style: NxTypography.captionMd.copyWith(color: context.nxColors.info)),
                      Align(
                        alignment: Alignment.centerRight,
                        child: NxButton(
                          label: 'Waste remove',
                          size: NxButtonSize.sm,
                          variant: NxButtonVariant.destructive,
                          onPressed: () => ref.read(expiryActionsProvider).wasteRemove(sku: item.sku, qty: item.qty),
                        ),
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
