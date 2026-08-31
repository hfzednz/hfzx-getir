import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/errors/customer_facing_error.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/stores_providers.dart';

class StoresScreen extends ConsumerWidget {
  const StoresScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final async = ref.watch(storesListProvider);
    final lang = Localizations.localeOf(context).languageCode;

    return Scaffold(
      appBar: NxTopBar(title: l10n.storesTitle),
      body: AsyncValueWidget(
        value: async,
        data: (stores) {
          if (stores.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(storesListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(storesListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: stores.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, i) {
                final store = stores[i];
                return ListTile(
                  contentPadding: const EdgeInsets.symmetric(
                    horizontal: NxSpacing.s3,
                    vertical: NxSpacing.s2,
                  ),
                  minVerticalPadding: 12,
                  leading: CircleAvatar(
                    child: Text(
                      store.name.isNotEmpty ? store.name[0].toUpperCase() : 'N',
                    ),
                  ),
                  title: Text(store.name),
                  subtitle: Text(
                    [
                      store.open ? l10n.storeOpen : l10n.storeClosed,
                      if (store.etaMinutes != null)
                        '${store.etaMinutes} ${l10n.deliveryEta}',
                      if (store.category != null) store.category!,
                    ].join(' · '),
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.push(
                    RouteNames.storeDetail.replaceFirst(':storeId', store.id),
                  ),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: customerFacingError(e, languageCode: lang),
          onRetry: () => ref.invalidate(storesListProvider),
        ),
      ),
    );
  }
}
