import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../../shared/utils/idempotency.dart';
import '../../../../shared/utils/money.dart';
import '../../domain/entities/checkout_entity.dart';
import '../providers/checkout_providers.dart';

class CheckoutReviewScreen extends ConsumerStatefulWidget {
  const CheckoutReviewScreen({super.key});

  @override
  ConsumerState<CheckoutReviewScreen> createState() =>
      _CheckoutReviewScreenState();
}

class _CheckoutReviewScreenState extends ConsumerState<CheckoutReviewScreen> {
  late final TextEditingController _couponController;
  bool _trackedReviewView = false;

  @override
  void initState() {
    super.initState();
    final code = ref.read(checkoutControllerProvider).couponCode;
    _couponController = TextEditingController(text: code ?? '');
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(checkoutControllerProvider.notifier).refreshQuote();
      if (!_trackedReviewView) {
        _trackedReviewView = true;
        // ignore: unawaited_futures
        ref.read(analyticsTrackerProvider).trackRaw(
              eventName: AnalyticsEvents.checkoutReviewViewed,
            );
      }
    });
  }

  @override
  void dispose() {
    _couponController.dispose();
    super.dispose();
  }

  String _amount(int minor) => Money(minorUnits: minor, currency: 'TRY')
      .majorUnits
      .toStringAsFixed(2);

  String _currencySymbol(String currency) => switch (currency.toUpperCase()) {
        'TRY' => '₺',
        'USD' => '\$',
        'EUR' => '€',
        _ => currency,
      };

  String _substitutionLabel(SubstitutionPreference pref) => switch (pref) {
        SubstitutionPreference.allow => 'Allow substitutions',
        SubstitutionPreference.contact => 'Contact before substituting',
        SubstitutionPreference.reject => 'Do not substitute',
      };

  String _oosLabel(OutOfStockReplacementRule rule) => switch (rule) {
        OutOfStockReplacementRule.similar => 'Replace with similar',
        OutOfStockReplacementRule.refund => 'Refund item',
        OutOfStockReplacementRule.cancel => 'Cancel order',
      };

  Future<void> _placeOrder() async {
    final ok = await ref.read(checkoutControllerProvider.notifier).placeOrder(
          idempotencyKey: Idempotency.generate(),
        );
    if (!mounted) return;
    if (!ok) {
      final message = ref.read(checkoutControllerProvider).errorMessage;
      if (message != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(message)),
        );
      }
      return;
    }

    final orderId = ref.read(checkoutControllerProvider).placedOrderId;
    if (orderId != null && orderId.isNotEmpty) {
      context.go('/orders/$orderId/track');
    } else {
      context.go(RouteNames.orders);
    }
  }

  Future<void> _retryPayment() async {
    final sessionId =
        ref.read(checkoutControllerProvider).lastPaymentSessionId;
    if (sessionId == null) return;
    final ok = await ref
        .read(checkoutControllerProvider.notifier)
        .retryPayment(sessionId);
    if (!mounted) return;
    if (!ok) {
      final message = ref.read(checkoutControllerProvider).errorMessage;
      if (message != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(message)),
        );
      }
      return;
    }

    final orderId = ref.read(checkoutControllerProvider).placedOrderId;
    if (orderId != null && orderId.isNotEmpty) {
      context.go('/orders/$orderId/track');
    } else {
      context.go(RouteNames.orders);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    final checkout = ref.watch(checkoutControllerProvider);
    final quote = checkout.quote;
    final canRetry = checkout.lastPaymentSessionId != null;

    return Scaffold(
      appBar: NxTopBar(title: l10n.checkoutTitle),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              children: [
                Text(
                  'Review your order',
                  style: NxTypography.headlineSm
                      .copyWith(color: colors.textPrimary),
                ),
                const SizedBox(height: NxSpacing.s4),
                if (checkout.isLoading && quote == null)
                  const Center(child: NxSpinner())
                else if (quote != null)
                  NxCheckoutSummary(
                    currencySymbol: _currencySymbol(quote.currency),
                    lines: [
                      NxCheckoutLine(
                        label: l10n.subtotal,
                        amount: _amount(quote.subtotalMinor),
                      ),
                      NxCheckoutLine(
                        label: l10n.deliveryFee,
                        amount: _amount(quote.deliveryFeeMinor),
                      ),
                      if (quote.discountMinor > 0)
                        NxCheckoutLine(
                          label: 'Discount',
                          amount: _amount(quote.discountMinor),
                          negative: true,
                        ),
                      if (quote.taxMinor > 0)
                        NxCheckoutLine(
                          label: 'Tax',
                          amount: _amount(quote.taxMinor),
                        ),
                      NxCheckoutLine(
                        label: l10n.total,
                        amount: _amount(quote.totalMinor),
                        emphasized: true,
                      ),
                    ],
                  )
                else
                  Text(
                    checkout.errorMessage ??
                        'Unable to load quote. Pull to refresh or try again.',
                    style: NxTypography.bodyMd
                        .copyWith(color: colors.textSecondary),
                  ),
                const SizedBox(height: NxSpacing.s5),
                Row(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Expanded(
                      child: NxField(
                        label: l10n.applyCoupon,
                        controller: _couponController,
                      ),
                    ),
                    const SizedBox(width: NxSpacing.s3),
                    NxButton(
                      label: l10n.applyCoupon,
                      variant: NxButtonVariant.secondary,
                      loading: checkout.isLoading,
                      onPressed: () => ref
                          .read(checkoutControllerProvider.notifier)
                          .applyCoupon(_couponController.text),
                    ),
                  ],
                ),
                if (checkout.errorMessage != null) ...[
                  const SizedBox(height: NxSpacing.s3),
                  Text(
                    checkout.errorMessage!,
                    style: NxTypography.bodySm.copyWith(color: colors.danger),
                  ),
                ],
                const SizedBox(height: NxSpacing.s4),
                Text(
                  checkout.scheduledAt == null
                      ? 'Delivery: ASAP'
                      : 'Delivery: ${checkout.scheduledAt!.toLocal()}',
                  style: NxTypography.bodySm
                      .copyWith(color: colors.textSecondary),
                ),
                if (checkout.contactless)
                  Text(
                    l10n.contactlessDelivery,
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                if (checkout.gift)
                  Text(
                    l10n.giftOrder,
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                const SizedBox(height: NxSpacing.s4),
                Text(
                  'Substitutions',
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                ),
                const SizedBox(height: NxSpacing.s1),
                Text(
                  _substitutionLabel(checkout.substitutionPreference),
                  style: NxTypography.bodySm
                      .copyWith(color: colors.textSecondary),
                ),
                Text(
                  _oosLabel(checkout.outOfStockRule),
                  style: NxTypography.bodySm
                      .copyWith(color: colors.textSecondary),
                ),
                if (checkout.wantInvoice) ...[
                  const SizedBox(height: NxSpacing.s3),
                  Text(
                    'Company invoice',
                    style:
                        NxTypography.titleSm.copyWith(color: colors.textPrimary),
                  ),
                  const SizedBox(height: NxSpacing.s1),
                  if (checkout.invoiceFields?.companyName != null)
                    Text(
                      checkout.invoiceFields!.companyName!,
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                  if (checkout.invoiceFields?.taxId != null)
                    Text(
                      'Tax ID: ${checkout.invoiceFields!.taxId}',
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                  if (checkout.invoiceFields?.taxOffice != null)
                    Text(
                      'Tax office: ${checkout.invoiceFields!.taxOffice}',
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                ],
                if (checkout.gift &&
                    checkout.giftMessage != null &&
                    checkout.giftMessage!.trim().isNotEmpty) ...[
                  const SizedBox(height: NxSpacing.s3),
                  Text(
                    'Gift message',
                    style:
                        NxTypography.titleSm.copyWith(color: colors.textPrimary),
                  ),
                  const SizedBox(height: NxSpacing.s1),
                  Text(
                    checkout.giftMessage!,
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
                ],
                if (checkout.installmentCount != null)
                  Padding(
                    padding: const EdgeInsets.only(top: NxSpacing.s3),
                    child: Text(
                      'Installments: ${checkout.installmentCount}x',
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                  ),
                if (checkout.paymentType == 'gift_card' &&
                    checkout.giftCardCode != null)
                  Padding(
                    padding: const EdgeInsets.only(top: NxSpacing.s2),
                    child: Text(
                      'Gift card applied',
                      style: NxTypography.bodySm
                          .copyWith(color: colors.textSecondary),
                    ),
                  ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                if (canRetry) ...[
                  NxButton(
                    label: l10n.retry,
                    variant: NxButtonVariant.secondary,
                    expand: true,
                    loading: checkout.isLoading,
                    onPressed: checkout.isLoading ? null : _retryPayment,
                  ),
                  const SizedBox(height: NxSpacing.s3),
                ],
                NxButton(
                  key: const ValueKey('place-order'),
                  label: l10n.placeOrder,
                  semanticLabel: 'Place order',
                  expand: true,
                  loading: checkout.isLoading,
                  onPressed: checkout.isLoading ? null : _placeOrder,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
