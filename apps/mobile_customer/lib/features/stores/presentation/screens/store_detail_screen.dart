import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/errors/customer_facing_error.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../search/presentation/providers/search_providers.dart';
import '../providers/stores_providers.dart';

class StoreDetailScreen extends ConsumerWidget {
  const StoreDetailScreen({super.key, required this.storeId});

  final String storeId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final lang = Localizations.localeOf(context).languageCode;
    final storeAsync = ref.watch(storeDetailProvider(storeId));
    final productsAsync = ref.watch(catalogBrowseProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.storesTitle),
      body: AsyncValueWidget(
        value: storeAsync,
        data: (store) {
          return RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(storeDetailProvider(storeId));
              ref.invalidate(catalogBrowseProvider);
            },
            child: ListView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              children: [
                Text(store.name, style: NxTypography.headlineSm),
                const SizedBox(height: NxSpacing.s2),
                Text(store.open ? l10n.storeOpen : l10n.storeClosed),
                if (store.etaMinutes != null)
                  Text('${l10n.deliveryEta}: ${store.etaMinutes} dk'),
                const SizedBox(height: NxSpacing.s4),
                Text(l10n.categoriesTitle, style: NxTypography.titleSm),
                const SizedBox(height: NxSpacing.s2),
                productsAsync.when(
                  loading: () => const Center(child: NxSpinner()),
                  error: (e, _) => ErrorView(
                    message: customerFacingError(e, languageCode: lang),
                    onRetry: () => ref.invalidate(catalogBrowseProvider),
                  ),
                  data: (items) {
                    if (items.isEmpty) {
                      return NxEmptyState(
                        title: l10n.emptyTitle,
                        body: l10n.emptyMessage,
                      );
                    }
                    return Column(
                      children: [
                        for (final item in items)
                          ListTile(
                            title: Text(item.title),
                            onTap: () => context.push(
                              RouteNames.product.replaceFirst(':productId', item.id),
                            ),
                          ),
                      ],
                    );
                  },
                ),
              ],
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: customerFacingError(e, languageCode: lang),
          onRetry: () => ref.invalidate(storeDetailProvider(storeId)),
        ),
      ),
    );
  }
}
