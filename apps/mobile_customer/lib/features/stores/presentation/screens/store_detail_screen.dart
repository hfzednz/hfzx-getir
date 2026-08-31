import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/errors/customer_facing_error.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../favorites/domain/entities/favorites_entity.dart';
import '../../../favorites/presentation/providers/favorites_providers.dart';
import '../../../product/domain/entities/product_entity.dart';
import '../providers/stores_providers.dart';

class StoreDetailScreen extends ConsumerWidget {
  const StoreDetailScreen({super.key, required this.storeId});

  final String storeId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final lang = Localizations.localeOf(context).languageCode;
    final storeAsync = ref.watch(storeDetailProvider(storeId));
    final productsAsync = ref.watch(storeProductsProvider(storeId));

    ref.listen(storeDetailProvider(storeId), (prev, next) {
      next.whenData((store) {
        ref.read(selectedStoreIdProvider.notifier).state = store.id;
      });
    });

    return Scaffold(
      appBar: NxTopBar(title: l10n.storesTitle),
      body: AsyncValueWidget(
        value: storeAsync,
        data: (store) {
          return RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(storeDetailProvider(storeId));
              ref.invalidate(storeProductsProvider(storeId));
            },
            child: ListView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              children: [
                Text(store.name, style: NxTypography.headlineSm),
                const SizedBox(height: NxSpacing.s2),
                Text(store.open ? l10n.storeOpen : l10n.storeClosed),
                if (!store.open) ...[
                  const SizedBox(height: NxSpacing.s2),
                  Text(l10n.storeClosedBody, style: NxTypography.bodyMd),
                ],
                if (store.etaMinutes != null)
                  Text('${l10n.deliveryEta}: ${store.etaMinutes} dk'),
                Text(
                  '${l10n.minOrder}: ${Formatters.money(Money(minorUnits: store.minOrderMinor, currency: 'TRY'))}',
                ),
                Text(
                  '${l10n.deliveryFee}: ${Formatters.money(Money(minorUnits: store.deliveryFeeMinor, currency: 'TRY'))}',
                ),
                const SizedBox(height: NxSpacing.s3),
                Align(
                  alignment: Alignment.centerLeft,
                  child: NxButton(
                    label: l10n.favorite,
                    variant: NxButtonVariant.secondary,
                    onPressed: () {
                      ref.read(favoritesRepositoryProvider).add(
                            FavoriteEntry(
                              id: 'store:$storeId',
                              type: FavoriteType.store,
                              targetId: storeId,
                              title: store.name,
                            ),
                          );
                    },
                  ),
                ),
                const SizedBox(height: NxSpacing.s4),
                Text(l10n.productTitle, style: NxTypography.titleSm),
                const SizedBox(height: NxSpacing.s2),
                productsAsync.when(
                  loading: () => const Center(child: NxSpinner()),
                  error: (e, _) => ErrorView(
                    message: customerFacingError(e, languageCode: lang),
                    onRetry: () => ref.invalidate(storeProductsProvider(storeId)),
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
                            contentPadding: EdgeInsets.zero,
                            minVerticalPadding: 12,
                            title: Text(item.title),
                            subtitle: Text(
                              item.stockStatus == ProductStockStatus.outOfStock
                                  ? l10n.outOfStock
                                  : Formatters.money(
                                      Money(
                                        minorUnits: item.priceMinor,
                                        currency: item.currency,
                                      ),
                                    ),
                            ),
                            enabled:
                                store.open &&
                                item.stockStatus != ProductStockStatus.outOfStock,
                            onTap: store.open &&
                                    item.stockStatus != ProductStockStatus.outOfStock
                                ? () => context.push(
                                      RouteNames.product.replaceFirst(
                                        ':productId',
                                        item.id,
                                      ),
                                    )
                                : null,
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
