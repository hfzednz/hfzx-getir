import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/home_entity.dart';
import '../providers/home_providers.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key, this.campaignSlug});

  final String? campaignSlug;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final feedAsync = ref.watch(homeFeedProvider);

    return Scaffold(
      appBar: NxTopBar(
        title: l10n.homeTitle,
        actions: [
          IconButton(
            icon: const Icon(Icons.auto_awesome_outlined),
            tooltip: 'AI picks',
            onPressed: () => context.push(RouteNames.ai),
          ),
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () => context.push(RouteNames.notifications),
          ),
        ],
      ),
      body: AsyncValueWidget(
        value: feedAsync,
        data: (feed) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(homeFeedProvider),
          child: CustomScrollView(
            slivers: [
              ...feed.widgets.map(
                (widget) => SliverToBoxAdapter(child: _HomeWidgetSection(config: widget)),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: 80)),
            ],
          ),
        ),
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate(homeFeedProvider),
        ),
      ),
    );
  }
}

class _HomeWidgetSection extends StatelessWidget {
  const _HomeWidgetSection({required this.config});

  final HomeWidgetConfig config;

  @override
  Widget build(BuildContext context) {
    return switch (config.type) {
      HomeWidgetType.banner => _BannerWidget(config: config),
      HomeWidgetType.countdown => _CountdownWidget(config: config),
      HomeWidgetType.flashSale => _ProductRailWidget(config: config, accent: true),
      HomeWidgetType.brands || HomeWidgetType.favoriteCategories => _ChipRailWidget(config: config),
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
        actionLabel: deepLink != null ? 'View' : null,
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
        message: endsAt != null ? 'Ends $endsAt' : 'Limited time offer',
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

class _ProductRailWidget extends StatelessWidget {
  const _ProductRailWidget({required this.config, this.accent = false});

  final HomeWidgetConfig config;
  final bool accent;

  @override
  Widget build(BuildContext context) {
    if (config.items.isEmpty) return const SizedBox.shrink();

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
                  onTap: () => context.push('/p/${p.id}'),
                  onAdd: () {},
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
