import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../../shared/errors/error_copy.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/cart_entity.dart';
import '../providers/cart_providers.dart';
import '../../../ai/presentation/providers/ai_providers.dart';

class CartScreen extends ConsumerStatefulWidget {
  const CartScreen({super.key});

  @override
  ConsumerState<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends ConsumerState<CartScreen> {
  final _couponController = TextEditingController();
  final _giftCardController = TextEditingController();
  final _walletController = TextEditingController();
  final _loyaltyController = TextEditingController();
  bool _tenderBusy = false;

  @override
  void dispose() {
    _couponController.dispose();
    _giftCardController.dispose();
    _walletController.dispose();
    _loyaltyController.dispose();
    super.dispose();
  }

  Future<void> _runTender(
    Future<Result<Cart>> Function() action, {
    String? couponCode,
  }) async {
    setState(() => _tenderBusy = true);
    try {
      final result = await action();
      await result.fold(
        onSuccess: (_) async {
          if (couponCode != null && couponCode.isNotEmpty) {
            await ref.read(analyticsTrackerProvider).trackRaw(
                  eventName: AnalyticsEvents.cartCouponApplied,
                  props: {'coupon_code': couponCode},
                );
          }
          ref.invalidate(cartEstimateProvider);
          ref.invalidate(cartItemsStreamProvider);
        },
        onFailure: (e) async {
          if (!mounted) return;
          NxToast.show(
            context,
            message: localizedCustomerError(context, e),
            variant: NxToastVariant.danger,
          );
        },
      );
    } finally {
      if (mounted) setState(() => _tenderBusy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final cartAsync = ref.watch(cartItemsStreamProvider);
    final estimateAsync = ref.watch(cartEstimateProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.cartTitle),
      body: AsyncValueWidget(
        value: cartAsync,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.homeTitle,
              onPrimaryAction: () => context.go(RouteNames.home),
            );
          }

          final cart = estimateAsync.maybeWhen(
            data: (c) => ref.read(cartRepositoryProvider).withRules(c),
            orElse: () {
              var subtotal = 0;
              for (final item in items) {
                subtotal += item.unitPriceMinor * item.quantity;
              }
              return ref.read(cartRepositoryProvider).withRules(
                    Cart(
                      id: 'local',
                      items: items
                          .map(
                            (i) => CartLine(
                              productId: i.productId,
                              variantId: i.variantId,
                              title: i.title,
                              imageUrl: i.imageUrl,
                              quantity: i.quantity,
                              unitPriceMinor: i.unitPriceMinor,
                              currency: i.currency,
                            ),
                          )
                          .toList(),
                      subtotalMinor: subtotal,
                      totalMinor: subtotal,
                      currency: items.first.currency,
                    ),
                  );
            },
          );

          final total = Money(minorUnits: cart.totalMinor, currency: cart.currency);
          final repo = ref.read(cartRepositoryProvider);

          return Column(
            children: [
              if (cart.violations.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.fromLTRB(
                    NxSpacing.s4,
                    NxSpacing.s4,
                    NxSpacing.s4,
                    0,
                  ),
                  child: Column(
                    children: cart.violations
                        .map(
                          (v) => Padding(
                            padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                            child: NxBanner(
                              variant: NxBannerVariant.danger,
                              message: v.message,
                            ),
                          ),
                        )
                        .toList(),
                  ),
                ),
              if (cart.promotions.isNotEmpty)
                Padding(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(l10n.promotions, style: NxTypography.headlineSm),
                      ...cart.promotions.map(
                        (promo) => ListTile(
                          dense: true,
                          title: Text(promo.title),
                          subtitle: promo.description != null ? Text(promo.description!) : null,
                          trailing: Text(
                            '-${Formatters.money(Money(minorUnits: promo.discountMinor, currency: cart.currency))}',
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              if (cart.coupon != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                  child: NxBanner(
                    variant: cart.coupon!.valid
                        ? NxBannerVariant.success
                        : NxBannerVariant.danger,
                    message: cart.coupon!.valid
                        ? '${l10n.couponApplied}: ${cart.coupon!.code}'
                        : '${l10n.couponInvalid}: ${cart.coupon!.code}',
                  ),
                ),
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  children: [
                    ...cart.items.map((item) {
                      final price = Money(
                        minorUnits: item.unitPriceMinor,
                        currency: item.currency,
                      );
                      return Padding(
                        padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                        child: NxCard(
                          child: Row(
                            children: [
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(item.title, style: NxTypography.bodyMd),
                                    NxPriceBlock(
                                      price: price.majorUnits.toStringAsFixed(2),
                                      currencySymbol: '₺',
                                    ),
                                    if (item.stockWarning != null)
                                      Text(
                                        item.stockWarning!,
                                        style: NxTypography.captionMd.copyWith(
                                          color: context.nxColors.warning,
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                              NxQtySelector(
                                quantity: item.quantity,
                                max: item.maxQty ?? 99,
                                semanticLabel: l10n.quantityLabel(item.quantity),
                                increaseSemanticLabel: l10n.increaseQuantity,
                                decreaseSemanticLabel: l10n.decreaseQuantity,
                                onIncrement: () => repo.updateQuantity(
                                  item.productId,
                                  item.quantity + 1,
                                  variantId: item.variantId,
                                ),
                                onDecrement: () => repo.updateQuantity(
                                  item.productId,
                                  item.quantity - 1,
                                  variantId: item.variantId,
                                ),
                              ),
                            ],
                          ),
                        ),
                      );
                    }),
                    const SizedBox(height: NxSpacing.s4),
                    Text(l10n.couponsAndPayments, style: NxTypography.headlineSm),
                    const SizedBox(height: NxSpacing.s3),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: NxField(
                            label: l10n.applyCoupon,
                            controller: _couponController,
                          ),
                        ),
                        const SizedBox(width: NxSpacing.s2),
                        NxButton(
                          label: l10n.apply,
                          variant: NxButtonVariant.secondary,
                          loading: _tenderBusy,
                          onPressed: () {
                            final code = _couponController.text.trim();
                            if (code.isEmpty) return;
                            _runTender(
                              () => repo.applyCoupon(code),
                              couponCode: code,
                            );
                          },
                        ),
                        if (cart.coupon != null) ...[
                          const SizedBox(width: NxSpacing.s2),
                          NxButton(
                            label: l10n.remove,
                            variant: NxButtonVariant.tertiary,
                            loading: _tenderBusy,
                            onPressed: () => _runTender(repo.removeCoupon),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: NxSpacing.s3),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: NxField(
                            label: l10n.giftCardCode,
                            controller: _giftCardController,
                          ),
                        ),
                        const SizedBox(width: NxSpacing.s2),
                        NxButton(
                          label: l10n.apply,
                          variant: NxButtonVariant.secondary,
                          loading: _tenderBusy,
                          onPressed: () {
                            final code = _giftCardController.text.trim();
                            if (code.isEmpty) return;
                            _runTender(() => repo.applyGiftCard(code));
                          },
                        ),
                      ],
                    ),
                    if (cart.giftCards.isNotEmpty) ...[
                      const SizedBox(height: NxSpacing.s2),
                      ...cart.giftCards.map(
                        (g) => Text(
                          '${l10n.giftCard} ${g.code}: -${Formatters.money(Money(minorUnits: g.appliedMinor, currency: cart.currency))}',
                          style: NxTypography.captionMd,
                        ),
                      ),
                    ],
                    const SizedBox(height: NxSpacing.s3),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: NxField(
                            label: l10n.walletAmount,
                            controller: _walletController,
                            keyboardType: const TextInputType.numberWithOptions(decimal: true),
                          ),
                        ),
                        const SizedBox(width: NxSpacing.s2),
                        NxButton(
                          label: l10n.apply,
                          variant: NxButtonVariant.secondary,
                          loading: _tenderBusy,
                          onPressed: () {
                            final major = double.tryParse(
                              _walletController.text.trim().replaceAll(',', '.'),
                            );
                            if (major == null || major <= 0) return;
                            final minor = (major * 100).round();
                            _runTender(() => repo.applyWallet(minor));
                          },
                        ),
                      ],
                    ),
                    if (cart.walletAppliedMinor > 0) ...[
                      const SizedBox(height: NxSpacing.s2),
                      Text(
                        '${l10n.walletApplied}: -${Formatters.money(Money(minorUnits: cart.walletAppliedMinor, currency: cart.currency))}',
                        style: NxTypography.captionMd,
                      ),
                    ],
                    const SizedBox(height: NxSpacing.s3),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Expanded(
                          child: NxField(
                            label: l10n.loyaltyRedeem,
                            controller: _loyaltyController,
                            keyboardType: TextInputType.number,
                          ),
                        ),
                        const SizedBox(width: NxSpacing.s2),
                        NxButton(
                          label: l10n.redeem,
                          variant: NxButtonVariant.secondary,
                          loading: _tenderBusy,
                          onPressed: () {
                            final points = int.tryParse(
                              _loyaltyController.text.trim(),
                            );
                            if (points == null || points <= 0) return;
                            _runTender(() => repo.applyLoyaltyPoints(points));
                          },
                        ),
                      ],
                    ),
                    if (cart.loyaltyPointsToRedeem > 0) ...[
                      const SizedBox(height: NxSpacing.s2),
                      Text(
                        '${l10n.loyaltyPointsApplied}: ${cart.loyaltyPointsToRedeem}',
                        style: NxTypography.captionMd,
                      ),
                    ],
                    const SizedBox(height: NxSpacing.s4),
                    Text(l10n.estimate, style: NxTypography.headlineSm),
                    const SizedBox(height: NxSpacing.s2),
                    _EstimateLine(
                      label: l10n.subtotal,
                      amount: Formatters.money(
                        Money(minorUnits: cart.subtotalMinor, currency: cart.currency),
                      ),
                    ),
                    if (cart.deliveryFeeEstimateMinor != null)
                      _EstimateLine(
                        label: l10n.deliveryFee,
                        amount: Formatters.money(
                          Money(
                            minorUnits: cart.deliveryFeeEstimateMinor!,
                            currency: cart.currency,
                          ),
                        ),
                      ),
                    if (cart.taxEstimateMinor != null)
                      _EstimateLine(
                        label: l10n.tax,
                        amount: Formatters.money(
                          Money(minorUnits: cart.taxEstimateMinor!, currency: cart.currency),
                        ),
                      ),
                    if (cart.coupon != null && cart.coupon!.discountMinor > 0)
                      _EstimateLine(
                        label: l10n.couponLabel,
                        amount:
                            '-${Formatters.money(Money(minorUnits: cart.coupon!.discountMinor, currency: cart.currency))}',
                      ),
                    _EstimateLine(
                      label: l10n.total,
                      amount: Formatters.money(total),
                      emphasized: true,
                    ),
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: l10n.validateInventory,
                      variant: NxButtonVariant.tertiary,
                      expand: true,
                      loading: _tenderBusy,
                      onPressed: () => _runTender(repo.validateInventory),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4, vertical: NxSpacing.s2),
                child: Row(
                  children: [
                    Expanded(
                      child: NxButton(
                        label: l10n.smartReorder,
                        variant: NxButtonVariant.secondary,
                        onPressed: () => ref.invalidate(aiReorderPredictionProvider),
                      ),
                    ),
                    const SizedBox(width: NxSpacing.s2),
                    Expanded(
                      child: NxButton(
                        label: l10n.budgetOptimize,
                        variant: NxButtonVariant.secondary,
                        onPressed: () => ref.read(aiRepositoryProvider).budgetOptimization(
                              budgetMinor: cart.totalMinor,
                            ),
                      ),
                    ),
                  ],
                ),
              ),
              if (cart.etaMinutes != null)
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                  child: Text(
                    '${l10n.deliveryEta}: ${Formatters.etaMinutes(cart.etaMinutes!)}',
                    style: NxTypography.captionMd,
                  ),
                ),
              NxCartBar(
                itemCount: cart.items.fold(0, (s, i) => s + i.quantity),
                total: total.majorUnits.toStringAsFixed(2),
                onViewCart: () {
                  if (!cart.canCheckout) {
                    NxToast.show(
                      context,
                      message: l10n.fixCartBeforeCheckout,
                      variant: NxToastVariant.danger,
                    );
                    return;
                  }
                  context.push(RouteNames.checkoutAddress);
                },
              ),
            ],
          );
        },
      ),
    );
  }
}

class _EstimateLine extends StatelessWidget {
  const _EstimateLine({
    required this.label,
    required this.amount,
    this.emphasized = false,
  });

  final String label;
  final String amount;
  final bool emphasized;

  @override
  Widget build(BuildContext context) {
    final style = emphasized ? NxTypography.headlineSm : NxTypography.bodyMd;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: NxSpacing.s1),
      child: Row(
        children: [
          Expanded(child: Text(label, style: style)),
          Text(amount, style: style),
        ],
      ),
    );
  }
}
