import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/checkout_providers.dart';

class CheckoutScreen extends ConsumerWidget {
  const CheckoutScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(checkoutListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.checkoutTitle),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(checkoutListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(checkoutListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                return NxCard(
                  child: ListTile(
                    title: Text(item.orderId ?? item.id),
                    subtitle: Text(item.status),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {},
                  ),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate(checkoutListProvider),
        ),
      ),
    );
  }
}
