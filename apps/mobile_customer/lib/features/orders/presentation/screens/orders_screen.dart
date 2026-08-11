import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/orders_providers.dart';

class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(ordersListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.ordersTitle),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(ordersListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(ordersListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                return NxOrderCard(
                  orderId: item.id,
                  statusLabel: item.statusLabel,
                  total: item.totalLabel,
                  subtitle: item.etaMinutes != null
                      ? 'ETA ~ ${item.etaMinutes} min'
                      : '${item.items.length} items',
                  imageUrls: item.items
                      .map((i) => i.imageUrl ?? '')
                      .where((u) => u.isNotEmpty)
                      .take(4)
                      .toList(),
                  extraItemCount:
                      item.items.length > 4 ? item.items.length - 4 : 0,
                  onTap: () => context.push('/orders/${item.id}'),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate(ordersListProvider),
        ),
      ),
    );
  }
}
