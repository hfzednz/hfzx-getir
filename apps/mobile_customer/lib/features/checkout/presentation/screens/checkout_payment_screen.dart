import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/checkout_providers.dart';

class CheckoutPaymentScreen extends ConsumerStatefulWidget {
  const CheckoutPaymentScreen({super.key});

  @override
  ConsumerState<CheckoutPaymentScreen> createState() =>
      _CheckoutPaymentScreenState();
}

class _CheckoutPaymentScreenState extends ConsumerState<CheckoutPaymentScreen> {
  static const _installmentOptions = [2, 3, 6, 9, 12];

  late final TextEditingController _giftCardController;
  late final TextEditingController _walletAmountController;

  @override
  void initState() {
    super.initState();
    final checkout = ref.read(checkoutControllerProvider);
    _giftCardController =
        TextEditingController(text: checkout.giftCardCode ?? '');
    final walletMajor = checkout.walletAmountMinor == null
        ? ''
        : (checkout.walletAmountMinor! / 100).toStringAsFixed(2);
    _walletAmountController = TextEditingController(text: walletMajor);
  }

  @override
  void dispose() {
    _giftCardController.dispose();
    _walletAmountController.dispose();
    super.dispose();
  }

  bool _isSelected({
    required String paymentType,
    required String? paymentMethodId,
    required String type,
    String? methodId,
  }) {
    if (paymentType != type) return false;
    if (type == 'card') return paymentMethodId == methodId;
    return paymentMethodId == type || paymentMethodId == methodId;
  }

  void _onWalletAmountChanged(String value) {
    final parsed = double.tryParse(value.replaceAll(',', '.'));
    if (parsed == null || parsed <= 0) {
      ref.read(checkoutControllerProvider.notifier).setWalletAmount(null);
      return;
    }
    final minor = (parsed * 100).round();
    ref.read(checkoutControllerProvider.notifier).setWalletAmount(minor);
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
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final colors = context.nxColors;
    final checkout = ref.watch(checkoutControllerProvider);
    final cardsAsync = ref.watch(paymentMethodsListProvider);
    final canContinue = checkout.paymentType == 'gift_card'
        ? (checkout.giftCardCode != null &&
            checkout.giftCardCode!.trim().isNotEmpty)
        : checkout.paymentType == 'card' ||
            checkout.paymentMethodId != null ||
            checkout.paymentType == 'wallet' ||
            checkout.paymentType == 'cash';

    return Scaffold(
      appBar: NxTopBar(title: l10n.paymentMethod),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Expanded(
            child: ListView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              children: [
                Text(
                  'How would you like to pay?',
                  style: NxTypography.headlineSm
                      .copyWith(color: colors.textPrimary),
                ),
                const SizedBox(height: NxSpacing.s4),
                _PaymentOptionTile(
                  selected: _isSelected(
                    paymentType: checkout.paymentType,
                    paymentMethodId: checkout.paymentMethodId,
                    type: 'card',
                    methodId: 'card',
                  ),
                  title: l10n.payWithCard,
                  subtitle: l10n.payWithCardHint,
                  icon: Icons.credit_card,
                  onTap: () => ref.read(checkoutControllerProvider.notifier).setPayment(
                        paymentType: 'card',
                        paymentMethodId: 'card',
                      ),
                ),
                const SizedBox(height: NxSpacing.s3),
                AsyncValueWidget(
                  value: cardsAsync,
                  loading: () => const Padding(
                    padding: EdgeInsets.symmetric(vertical: NxSpacing.s4),
                    child: Center(child: NxSpinner()),
                  ),
                  data: (cards) {
                    if (cards.isEmpty) {
                      return Padding(
                        padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                        child: Text(
                          'No saved cards yet. You can still pay with wallet or cash.',
                          style: NxTypography.bodySm
                              .copyWith(color: colors.textSecondary),
                        ),
                      );
                    }
                    return Column(
                      children: [
                        for (final card in cards) ...[
                          _PaymentOptionTile(
                            selected: _isSelected(
                              paymentType: checkout.paymentType,
                              paymentMethodId: checkout.paymentMethodId,
                              type: 'card',
                              methodId: card.id,
                            ),
                            title: '${card.brand} ···· ${card.last4}',
                            subtitle: card.expiryMonth != null &&
                                    card.expiryYear != null
                                ? 'Expires ${card.expiryMonth}/${card.expiryYear}'
                                : (card.isDefault ? 'Default card' : null),
                            icon: Icons.credit_card,
                            onTap: () => ref
                                .read(checkoutControllerProvider.notifier)
                                .setPayment(
                                  paymentType: 'card',
                                  paymentMethodId: card.id,
                                ),
                          ),
                          const SizedBox(height: NxSpacing.s3),
                        ],
                      ],
                    );
                  },
                  error: (e, _) => Padding(
                    padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                    child: ErrorView(
                      message: e.toString(),
                      onRetry: () =>
                          ref.invalidate(paymentMethodsListProvider),
                    ),
                  ),
                ),
                if (checkout.paymentType == 'card') ...[
                  Text(
                    'Installments',
                    style: NxTypography.titleSm
                        .copyWith(color: colors.textPrimary),
                  ),
                  const SizedBox(height: NxSpacing.s2),
                  Wrap(
                    spacing: NxSpacing.s2,
                    runSpacing: NxSpacing.s2,
                    children: [
                      FilterChip(
                        label: const Text('Pay in full'),
                        selected: checkout.installmentCount == null,
                        onSelected: (_) => ref
                            .read(checkoutControllerProvider.notifier)
                            .setInstallments(null),
                      ),
                      for (final count in _installmentOptions)
                        FilterChip(
                          label: Text('${count}x'),
                          selected: checkout.installmentCount == count,
                          onSelected: (_) => ref
                              .read(checkoutControllerProvider.notifier)
                              .setInstallments(count),
                        ),
                    ],
                  ),
                  const SizedBox(height: NxSpacing.s4),
                ],
                _PaymentOptionTile(
                  selected: _isSelected(
                    paymentType: checkout.paymentType,
                    paymentMethodId: checkout.paymentMethodId,
                    type: 'wallet',
                  ),
                  title: l10n.walletTitle,
                  subtitle: 'Pay from your NEXORA wallet balance',
                  icon: Icons.account_balance_wallet_outlined,
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setPayment(
                        paymentType: 'wallet',
                        paymentMethodId: 'wallet',
                      ),
                ),
                const SizedBox(height: NxSpacing.s3),
                NxField(
                  label: 'Wallet amount (optional)',
                  controller: _walletAmountController,
                  keyboardType:
                      const TextInputType.numberWithOptions(decimal: true),
                  onChanged: _onWalletAmountChanged,
                ),
                const SizedBox(height: NxSpacing.s3),
                _PaymentOptionTile(
                  selected: _isSelected(
                    paymentType: checkout.paymentType,
                    paymentMethodId: checkout.paymentMethodId,
                    type: 'cash',
                  ),
                  title: l10n.cashOnDelivery,
                  subtitle: 'Pay the courier when your order arrives',
                  icon: Icons.payments_outlined,
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setPayment(paymentType: 'cash', paymentMethodId: 'cash'),
                ),
                const SizedBox(height: NxSpacing.s3),
                _PaymentOptionTile(
                  selected: _isSelected(
                    paymentType: checkout.paymentType,
                    paymentMethodId: checkout.paymentMethodId,
                    type: 'gift_card',
                  ),
                  title: 'Gift card',
                  subtitle: 'Redeem a gift card balance',
                  icon: Icons.card_giftcard_outlined,
                  onTap: () => ref
                      .read(checkoutControllerProvider.notifier)
                      .setPayment(
                        paymentType: 'gift_card',
                        paymentMethodId: 'gift_card',
                      ),
                ),
                if (checkout.paymentType == 'gift_card') ...[
                  const SizedBox(height: NxSpacing.s3),
                  NxField(
                    label: 'Gift card code',
                    controller: _giftCardController,
                    onChanged: (value) => ref
                        .read(checkoutControllerProvider.notifier)
                        .setGiftCard(value),
                  ),
                ],
                if (checkout.lastPaymentSessionId != null) ...[
                  const SizedBox(height: NxSpacing.s5),
                  Text(
                    checkout.errorMessage ??
                        'Previous payment failed. You can retry.',
                    style: NxTypography.bodySm.copyWith(color: colors.danger),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  NxButton(
                    label: l10n.retry,
                    variant: NxButtonVariant.secondary,
                    expand: true,
                    loading: checkout.isLoading,
                    onPressed: checkout.isLoading ? null : _retryPayment,
                  ),
                ],
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: NxButton(
              label: l10n.continueLabel,
              expand: true,
              disabled: !canContinue,
              onPressed: !canContinue
                  ? null
                  : () {
                      ref.read(checkoutControllerProvider.notifier).refreshQuote();
                      context.push(RouteNames.checkoutReview);
                    },
            ),
          ),
        ],
      ),
    );
  }
}

class _PaymentOptionTile extends StatelessWidget {
  const _PaymentOptionTile({
    required this.selected,
    required this.title,
    required this.icon,
    required this.onTap,
    this.subtitle,
  });

  final bool selected;
  final String title;
  final String? subtitle;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return NxCard(
      onTap: onTap,
      child: Row(
        children: [
          Icon(
            selected ? Icons.radio_button_checked : Icons.radio_button_off,
            color: selected ? colors.bgBrand : colors.textTertiary,
          ),
          const SizedBox(width: NxSpacing.s3),
          Icon(icon, color: colors.textSecondary),
          const SizedBox(width: NxSpacing.s3),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                ),
                if (subtitle != null)
                  Text(
                    subtitle!,
                    style: NxTypography.bodySm
                        .copyWith(color: colors.textSecondary),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
