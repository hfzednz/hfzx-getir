import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/transfer_entity.dart';
import '../providers/transfers_providers.dart';

class TransfersScreen extends ConsumerWidget {
  const TransfersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(transfersListProvider);
    return Scaffold(
      appBar: NxTopBar(
        title: 'Transfers',
        actions: [
          IconButton(icon: const Icon(Icons.add), onPressed: () => context.push(RouteNames.transferCreate)),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(transfersListProvider),
        child: AsyncValueWidget<List<WarehouseTransfer>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [SizedBox(height: 120), NxEmptyState(title: 'No transfers')]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final t = items[i];
                return NxCard(
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('${t.type} · ${t.fromLocation} → ${t.toLocation}', style: NxTypography.titleSm),
                            Text('${t.sku ?? '-'} × ${t.qty} · ${t.status.name}', style: NxTypography.captionMd),
                          ],
                        ),
                      ),
                      if (t.status == TransferStatus.pending)
                        NxButton(
                          label: 'Approve',
                          size: NxButtonSize.sm,
                          onPressed: () => ref.read(transfersActionsProvider).approve(t.id),
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
