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
import '../../../cart/presentation/providers/cart_providers.dart';
import '../../../orders/domain/entities/orders_entity.dart';
import '../../../orders/presentation/providers/orders_providers.dart';
import '../../domain/entities/home_entity.dart';
import '../providers/home_providers.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key, this.campaignSlug});

  final String? campaignSlug;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final lang = Localizations.localeOf(context).languageCode;
    final feedAsync = ref.watch(homeFeedProvider);
    final cartCount = ref.watch(cartItemCountProvider);
    final ordersAsync = ref.watch(ordersListProvider);

    return Scaffold(
      appBar: NxTopBar(
        title: l10n.homeTitle,
        actions: [
          IconButton(
            icon: const Icon(Icons.storefront_outlined),
            tooltip: l10n.storesTitle,
            onPressed: () => context.push(RouteNames.stores),
          ),
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            tooltip: l10n.notificationsTitle,
            onPressed: () => context.push(RouteNames.notifications),
          ),
        ],
      ),
      body: AsyncValueWidget(
        value: feedAsync,
        data: (feed) => RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(homeFeedProvider);
            ref.invalidate(ordersListProvider);
          },
          child: CustomScrollView(
            slivers: [
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(
                    NxSpacing.s4,
                    NxSpacing.s3,
                    NxSpacing.s4,
                    NxSpacing.s2,
                  ),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(12),
                    onTap: () => context.push(RouteNames.search),
                    child: InputDecorator(
                      decoration: InputDecoration(
                        hintText: l10n.searchHint,
                        prefixIcon: const Icon(Icons.search),
                        border: const OutlineInputBorder(),
                      ),
                      child: const SizedBox(height: 20),
                    ),
                  ),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                  child: NxBanner(
                    title: feed.serviceable
                        ? l10n.storesTitle
                        : l10n.emptyTitle,
                    message: feed.serviceable
                        ? l10n.deliveryEta
                        : l10n.emptyMessage,
                  ),
                ),
              ),
              ..._activeOrderSlivers(context, l10n, ordersAsync),
              if (cartCount > 0)
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.fromLTRB(
                      NxSpacing.s4,
                      NxSpacing.s3,
                      NxSpacing.s4,
                      0,
                    ),
                    child: NxBanner(
                      title: l10n.cartTitle,
                      message: '$cartCount',
                      actionLabel: l10n.checkout,
                      onAction: () => context.go(RouteNames.cart),
                    ),
                  ),
                ),
              if (feed.widgets.isEmpty)
                SliverFillRemaining(
                  hasScrollBody: false,
                  child: NxEmptyState(
                    title: l10n.emptyTitle,
                    body: l10n.emptyMessage,
                    primaryActionLabel: l10n.retry,
                    onPrimaryAction: () => ref.invalidate(homeFeedProvider),
                  ),
                )
              else
                ...feed.widgets.map(
                  (widget) => SliverToBoxAdapter(
                    child: _HomeWidgetSection(config: widget),
                  ),
                ),
              const SliverToBoxAdapter(child: SizedBox(height: 80)),
            ],
          ),
        ),
        error: (e, _) => ErrorView(
          message: customerFacingError(e, languageCode: lang),
          onRetry: () => ref.invalidate(homeFeedProvider),
        ),
      ),
    );
  }
}

List<Widget> _activeOrderSlivers(
  BuildContext context,
  AppLocalizations l10n,
  AsyncValue<List<Order>> ordersAsync,
) {
  return ordersAsync.maybeWhen(
    data: (orders) {
      final active = orders.where((o) {
        final name = o.status.name;
        return name != 'delivered' &&
            name != 'cancelled' &&
            name != 'refunded';
      }).toList();
      if (active.isEmpty) return const <Widget>[];
      final order = active.first;
      return [
        SliverToBoxAdapter(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(
              NxSpacing.s4,
              NxSpacing.s3,
              NxSpacing.s4,
              0,
            ),
            child: NxBanner(
              title: l10n.trackingTitle,
              message: order.id,
              actionLabel: l10n.continueLabel,
              onAction: () => context.push(
                RouteNames.orderTrack.replaceFirst(':orderId', order.id),
              ),
            ),
          ),
        ),
      ];
    },
    orElse: () => const <Widget>[],
  );
}

class _HomeWidgetSection extends StatelessWidget {
  const _HomeWidgetSection({required this.config});

  final HomeWidgetConfig config;

  @override
  Widget build(BuildContext context) {
    return switch (config.type) {
      HomeWidgetType.banner ||
      HomeWidgetType.campaign =>
        _BannerWidget(config: config),
      HomeWidgetType.countdown => _CountdownWidget(config: config),
      HomeWidgetType.flashSale => _ProductRailWidget(config: config, accent: true),
      HomeWidgetType.brands ||
      HomeWidgetType.favoriteCategories =>
        _ChipRailWidget(config: config),
      _ => _ProductRailWidget(config: config),
    };
  }
}

class _BannerWidget extends StatelessWidget {
  const _BannerWidget({required this.config});

  final HomeWidgetConfig config;

  @override
  Widget build(BuildContext context) {
    final title =
        config.title.isNotEmpty ? config.title : config.payload['title']?.toString() ?? '';
    final subtitle = config.payload['subtitle']?.toString() ?? '';
    final deepLink = config.payload['deep_link']?.toString();

    return Padding(
      padding: const EdgeInsets.all(NxSpacing.s4),
      child: NxBanner(
        title: title,
        message: subtitle,
        onAction: deepLink != null ? () => context.push(deepLink) : null,
        actionLabel: deepLink != null ? AppLocalizations.of(context).continueLabel : null,
      ),
    );
  }
}

class _CountdownWidget extends StatelessWidget {
  const _CountdownWidget({required this.config});

  final HomeWidgetConfig config;

  @override
  Widget build(BuildContext context) {
    final endsAt = config.payload['ends_at']?.toString();
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4, vertical: NxSpacing.s2),
      child: NxBanner(
        title: config.title,
        message: endsAt ?? AppLocalizations.of(context).deliveryEta,
      ),
    );
  }
}

class _ChipRailWidget extends StatelessWidget {
  const _ChipRailWidget({required this.config});

  final HomeWidgetConfig config;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (config.title.isNotEmpty)
          Padding(
            padding: const EdgeInsets.fromLTRB(
              NxSpacing.s4,
              NxSpacing.s4,
              NxSpacing.s4,
              NxSpacing.s2,
            ),
            child: Text(config.title, style: NxTypography.headlineSm),
          ),
        SizedBox(
          height: 48,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
            itemCount: config.items.length,
            separatorBuilder: (_, __) => const SizedBox(width: NxSpacing.s2),
            itemBuilder: (context, i) {
              final item = config.items[i];
              return NxChip(
                label: item.title,
                onSelected: (_) {
                  final link = item.deepLink ?? '/categories/${item.id}';
                  context.push(link);
                },
              );
            },
          ),
        ),
      ],
    );
  }
}

class _ProductRailWidget extends ConsumerWidget {
  const _ProductRailWidget({required this.config, this.accent = false});

  final HomeWidgetConfig config;
  final bool accent;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    if (config.items.isEmpty) return const SizedBox.shrink();
    final l10n = AppLocalizations.of(context);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(
            NxSpacing.s4,
            NxSpacing.s4,
            NxSpacing.s4,
            NxSpacing.s2,
          ),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  config.title,
                  style: NxTypography.headlineSm.copyWith(
                    color: accent ? context.nxColors.textBrand : null,
                  ),
                ),
              ),
              if (accent) const Icon(Icons.bolt, size: 18),
            ],
          ),
        ),
        SizedBox(
          height: 220,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
            itemCount: config.items.length,
            separatorBuilder: (_, __) => const SizedBox(width: NxSpacing.s3),
            itemBuilder: (context, index) {
              final p = config.items[index];
              final price = Money(minorUnits: p.priceMinor, currency: p.currency);
              return SizedBox(
                width: 150,
                child: NxProductCard(
                  title: p.title,
                  price: Formatters.money(price),
                  imageUrl: p.imageUrl,
                  unitMeta: p.unitMeta,
                  onTap: () => context.push(
                    p.deepLink ?? RouteNames.product.replaceFirst(':productId', p.id),
                  ),
                  onAdd: () {
                    ref.read(cartRepositoryProvider).addItem(
                          productId: p.id,
                          title: p.title,
                          imageUrl: p.imageUrl,
                          unitPriceMinor: p.priceMinor,
                          currency: p.currency,
                        ).then((_) {
                      if (context.mounted) {
                        NxToast.show(context, message: l10n.addToCart);
                      }
                    });
                  },
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
