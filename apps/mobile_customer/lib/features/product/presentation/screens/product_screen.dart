import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:share_plus/share_plus.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../cart/presentation/providers/cart_providers.dart';
import '../../domain/entities/product_entity.dart';
import '../providers/product_providers.dart';
import '../../../ai/presentation/providers/ai_providers.dart';

class ProductScreen extends ConsumerStatefulWidget {
  const ProductScreen({super.key, required this.productId});

  final String productId;

  @override
  ConsumerState<ProductScreen> createState() => _ProductScreenState();
}

class _ProductScreenState extends ConsumerState<ProductScreen> {
  int _qty = 1;
  String? _selectedVariantId;
  final _questionController = TextEditingController();
  bool _asking = false;
  bool _trackedView = false;

  @override
  void dispose() {
    _questionController.dispose();
    super.dispose();
  }

  void _trackViewOnce(Product p) {
    if (_trackedView) return;
    _trackedView = true;
    // ignore: unawaited_futures
    ref.read(analyticsTrackerProvider).trackProductViewed(
          productId: p.id,
          priceMinor: p.priceMinor,
          currency: p.currency,
        );
  }

  Future<void> _share(Product p) async {
    await Share.share(
      '${p.title}\n/p/${p.id}',
      subject: p.title,
    );
  }

  Future<void> _askQuestion(Product p) async {
    final text = _questionController.text.trim();
    if (text.isEmpty) return;
    setState(() => _asking = true);
    final result = await ref.read(productRepositoryProvider).askQuestion(
          productId: p.id,
          question: text,
        );
    if (!mounted) return;
    setState(() => _asking = false);
    await result.fold(
      onSuccess: (_) async {
        _questionController.clear();
        ref.invalidate(productDetailProvider(widget.productId));
        NxToast.show(context, message: 'Question submitted', variant: NxToastVariant.success);
      },
      onFailure: (e) async {
        NxToast.show(context, message: e.message, variant: NxToastVariant.danger);
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final productAsync = ref.watch(productDetailProvider(widget.productId));
    final priceHistoryAsync = ref.watch(productPriceHistoryProvider(widget.productId));

    return Scaffold(
      body: productAsync.when(
        data: (p) {
          if (!_trackedView) {
            WidgetsBinding.instance.addPostFrameCallback((_) => _trackViewOnce(p));
          }
          final priceMinor = _selectedVariantId != null
              ? p.variants
                      .where((v) => v.id == _selectedVariantId)
                      .map((v) => v.priceMinor)
                      .firstOrNull ??
                  p.priceMinor
              : p.priceMinor;
          final price = Money(minorUnits: priceMinor, currency: p.currency);
          final compareAt = p.compareAtMinor != null
              ? Money(minorUnits: p.compareAtMinor!, currency: p.currency)
              : null;

          return CustomScrollView(
            slivers: [
              SliverAppBar(
                expandedHeight: 280,
                pinned: true,
                actions: [
                  IconButton(
                    icon: const Icon(Icons.share_outlined),
                    onPressed: () => _share(p),
                  ),
                  IconButton(
                    icon: Icon(p.isFavorite ? Icons.favorite : Icons.favorite_border),
                    onPressed: () async {
                      await ref.read(productRepositoryProvider).toggleFavorite(p.id);
                      ref.invalidate(productDetailProvider(widget.productId));
                    },
                  ),
                ],
                flexibleSpace: FlexibleSpaceBar(
                  background: p.primaryImageUrl != null
                      ? Image.network(p.primaryImageUrl!, fit: BoxFit.cover)
                      : ColoredBox(color: context.nxColors.bgSurfaceRaised),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      if (p.brand != null) ...[
                        Text(p.brand!, style: NxTypography.captionMd),
                        const SizedBox(height: NxSpacing.s1),
                      ],
                      Text(p.title, style: NxTypography.headlineMd),
                      const SizedBox(height: NxSpacing.s2),
                      NxPriceBlock(
                        price: price.majorUnits.toStringAsFixed(2),
                        originalPrice: compareAt?.majorUnits.toStringAsFixed(2),
                        currencySymbol: '₺',
                      ),
                      if (p.discountPercent != null && p.discountPercent! > 0) ...[
                        const SizedBox(height: NxSpacing.s1),
                        Text(
                          '${p.discountPercent}% off',
                          style: NxTypography.captionMd.copyWith(
                            color: context.nxColors.textBrand,
                          ),
                        ),
                      ],
                      const SizedBox(height: NxSpacing.s2),
                      NxStockIndicator(
                        status: toNxStockStatus(p.stockStatus),
                        lowStockLabel: p.lowStockThreshold != null
                            ? 'Only ${p.lowStockThreshold} left'
                            : null,
                      ),
                      if (p.dietaryTags.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s3),
                        Wrap(
                          spacing: NxSpacing.s2,
                          children: p.dietaryTags.map((t) => NxChip(label: t)).toList(),
                        ),
                      ],
                      if (p.variants.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text('Variants', style: NxTypography.headlineSm),
                        const SizedBox(height: NxSpacing.s2),
                        Wrap(
                          spacing: NxSpacing.s2,
                          children: p.variants.map((v) {
                            final selected = _selectedVariantId == v.id;
                            return NxChip(
                              label: v.label,
                              selected: selected,
                              onSelected: (_) => setState(() => _selectedVariantId = v.id),
                            );
                          }).toList(),
                        ),
                      ],
                      const SizedBox(height: NxSpacing.s4),
                      Row(
                        children: [
                          Text('Quantity', style: NxTypography.bodyMd),
                          const Spacer(),
                          NxQtySelector(
                            quantity: _qty,
                            max: p.maxQty ?? 99,
                            onIncrement: () => setState(() => _qty++),
                            onDecrement: () => setState(() {
                              if (_qty > 1) _qty--;
                            }),
                          ),
                        ],
                      ),
                      if (p.description != null) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text('Description', style: NxTypography.headlineSm),
                        const SizedBox(height: NxSpacing.s2),
                        Text(p.description!, style: NxTypography.bodyMd),
                      ],
                      if (p.ingredients.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _SectionList(title: 'Ingredients', items: p.ingredients),
                      ],
                      if (p.allergens.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _SectionList(title: 'Allergens', items: p.allergens),
                      ],
                      if (p.origin != null) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text('Origin: ${p.origin}', style: NxTypography.bodyMd),
                      ],
                      if (p.nutrition != null) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text('Nutrition', style: NxTypography.headlineSm),
                        const SizedBox(height: NxSpacing.s2),
                        if (p.nutrition!.servingSize != null)
                          Text('Serving: ${p.nutrition!.servingSize}'),
                        if (p.nutrition!.calories != null)
                          Text('Calories: ${p.nutrition!.calories}'),
                        const SizedBox(height: NxSpacing.s2),
                        NxButton(
                          label: 'AI nutrition tips',
                          variant: NxButtonVariant.secondary,
                          onPressed: () => ref
                              .read(aiRepositoryProvider)
                              .nutritionSuggestions(productIds: [p.id]),
                        ),
                      ],
                      priceHistoryAsync.when(
                        data: (points) {
                          if (points.isEmpty) return const SizedBox.shrink();
                          return Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const SizedBox(height: NxSpacing.s4),
                              Text('Price history', style: NxTypography.headlineSm),
                              const SizedBox(height: NxSpacing.s2),
                              ...points.take(8).map((point) {
                                final m = Money(
                                  minorUnits: point.priceMinor,
                                  currency: point.currency,
                                );
                                final date =
                                    '${point.date.year}-${point.date.month.toString().padLeft(2, '0')}-${point.date.day.toString().padLeft(2, '0')}';
                                return ListTile(
                                  dense: true,
                                  contentPadding: EdgeInsets.zero,
                                  title: Text(date),
                                  trailing: Text(Formatters.money(m)),
                                );
                              }),
                            ],
                          );
                        },
                        loading: () => const SizedBox.shrink(),
                        error: (_, __) => const SizedBox.shrink(),
                      ),
                      if (p.bundles.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text('Bundles', style: NxTypography.headlineSm),
                        const SizedBox(height: NxSpacing.s2),
                        ...p.bundles.map((bundle) {
                          final bundlePrice = Money(
                            minorUnits: bundle.priceMinor,
                            currency: bundle.currency,
                          );
                          return Padding(
                            padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                            child: NxCard(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(bundle.title, style: NxTypography.bodyMd),
                                  Text(Formatters.money(bundlePrice), style: NxTypography.captionMd),
                                  if (bundle.items.isNotEmpty)
                                    Text(
                                      bundle.items
                                          .map((i) => '${i.title ?? i.productId} ×${i.quantity}')
                                          .join(', '),
                                      style: NxTypography.captionMd,
                                    ),
                                ],
                              ),
                            ),
                          );
                        }),
                      ],
                      if (p.reviewsSummary != null && p.reviewsSummary!.count > 0) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text(
                          'Reviews ${p.reviewsSummary!.averageRating.toStringAsFixed(1)} (${p.reviewsSummary!.count})',
                          style: NxTypography.headlineSm,
                        ),
                      ],
                      if (p.qaSummary != null && p.qaSummary!.questionCount > 0) ...[
                        const SizedBox(height: NxSpacing.s4),
                        Text(
                          'Q&A: ${p.qaSummary!.questionCount} questions',
                          style: NxTypography.bodyMd,
                        ),
                      ],
                      const SizedBox(height: NxSpacing.s4),
                      Text('Questions & answers', style: NxTypography.headlineSm),
                      const SizedBox(height: NxSpacing.s2),
                      if (p.questions.isNotEmpty)
                        ...p.questions.map(
                          (q) => Padding(
                            padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text('Q: ${q.question}', style: NxTypography.bodyMd),
                                if (q.answer != null)
                                  Text(
                                    'A: ${q.answer}',
                                    style: NxTypography.captionMd.copyWith(
                                      color: context.nxColors.textSecondary,
                                    ),
                                  ),
                              ],
                            ),
                          ),
                        )
                      else
                        Text(
                          'No questions yet. Be the first to ask.',
                          style: NxTypography.captionMd,
                        ),
                      const SizedBox(height: NxSpacing.s3),
                      NxField(
                        label: 'Ask a question',
                        controller: _questionController,
                      ),
                      const SizedBox(height: NxSpacing.s2),
                      NxButton(
                        label: 'Submit question',
                        variant: NxButtonVariant.secondary,
                        loading: _asking,
                        onPressed: () => _askQuestion(p),
                      ),
                      if (p.alternatives.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _ProductRail(title: 'Alternatives', products: p.alternatives),
                      ],
                      if (p.crossSell.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _ProductRail(title: 'Frequently bought together', products: p.crossSell),
                      ],
                      if (p.upsell.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _ProductRail(title: 'You may also like', products: p.upsell),
                      ],
                      if (p.aiRecommendations.isNotEmpty) ...[
                        const SizedBox(height: NxSpacing.s4),
                        _ProductRail(title: 'Recommended for you', products: p.aiRecommendations),
                      ],
                      const SizedBox(height: NxSpacing.s6),
                      NxButton(
                        label: l10n.addToCart,
                        expand: true,
                        onPressed: p.stockStatus == ProductStockStatus.outOfStock
                            ? null
                            : () => ref.read(cartRepositoryProvider).addItem(
                                  productId: p.id,
                                  variantId: _selectedVariantId,
                                  title: p.title,
                                  imageUrl: p.primaryImageUrl,
                                  unitPriceMinor: priceMinor,
                                  currency: p.currency,
                                  quantity: _qty,
                                ),
                      ),
                      const SizedBox(height: NxSpacing.s8),
                    ],
                  ),
                ),
              ),
            ],
          );
        },
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}

class _SectionList extends StatelessWidget {
  const _SectionList({required this.title, required this.items});

  final String title;
  final List<String> items;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: NxTypography.headlineSm),
        const SizedBox(height: NxSpacing.s2),
        Text(items.join(', '), style: NxTypography.bodyMd),
      ],
    );
  }
}

class _ProductRail extends StatelessWidget {
  const _ProductRail({required this.title, required this.products});

  final String title;
  final List<ProductSummary> products;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: NxTypography.headlineSm),
        const SizedBox(height: NxSpacing.s2),
        SizedBox(
          height: 200,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: products.length,
            separatorBuilder: (_, __) => const SizedBox(width: NxSpacing.s3),
            itemBuilder: (context, i) {
              final item = products[i];
              final price = Money(minorUnits: item.priceMinor, currency: item.currency);
              return SizedBox(
                width: 140,
                child: NxProductCard(
                  title: item.title,
                  price: Formatters.money(price),
                  imageUrl: item.imageUrl,
                  stockStatus: toNxStockStatus(item.stockStatus),
                  onTap: () => context.push('/p/${item.id}'),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
