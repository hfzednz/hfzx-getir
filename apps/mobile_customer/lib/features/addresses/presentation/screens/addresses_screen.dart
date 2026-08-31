import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/addresses_providers.dart';
import '../../../../shared/errors/error_copy.dart';

class AddressesScreen extends ConsumerWidget {
  const AddressesScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(addressesListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.addressesTitle),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(addressesListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(addressesListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                return NxCard(
                  child: ListTile(
                    title: Text(item.title.isNotEmpty ? item.title : item.id),
                    subtitle: Text(item.id),
                    trailing: const Icon(Icons.chevron_right),
                    onTap: () {},
                  ),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: localizedCustomerError(context, e),
          onRetry: () => ref.invalidate(addressesListProvider),
        ),
      ),
    );
  }
}
