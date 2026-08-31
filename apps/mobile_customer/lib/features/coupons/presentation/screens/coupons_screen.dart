import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/business_rules/coupon_rules.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/coupons_entity.dart';
import '../providers/coupons_providers.dart';
import '../../../../shared/errors/error_copy.dart';

class CouponsScreen extends ConsumerStatefulWidget {
  const CouponsScreen({super.key, this.promoCode});

  final String? promoCode;

  @override
  ConsumerState<CouponsScreen> createState() => _CouponsScreenState();
}

class _CouponsScreenState extends ConsumerState<CouponsScreen> {
  final _codeController = TextEditingController();

  @override
  void initState() {
    super.initState();
    if (widget.promoCode != null) _codeController.text = widget.promoCode!;
  }

  @override
  void dispose() {
    _codeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(couponsListProvider);
    final applyState = ref.watch(couponApplyControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.couponsTitle),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: _codeController,
                    decoration: InputDecoration(
                      labelText: l10n.applyCoupon,
                      border: const OutlineInputBorder(),
                    ),
                    textCapitalization: TextCapitalization.characters,
                  ),
                ),
                const SizedBox(width: NxSpacing.s2),
                NxButton(
                  label: l10n.apply,
                  loading: applyState.isLoading,
                  onPressed: () => ref.read(couponApplyControllerProvider.notifier).apply(
                        code: _codeController.text,
                        cartSubtotalMinor: 0,
                        cartCurrency: 'TRY',
                      ),
                ),
              ],
            ),
          ),
          if (applyState.hasValue && applyState.value != null)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
              child: NxBanner(
                title: l10n.couponApplied,
                message: applyState.value!.message.isNotEmpty
                    ? applyState.value!.message
                    : '${applyState.value!.coupon.code} — ${Money(minorUnits: applyState.value!.discountMinor, currency: 'TRY').format()} off',
                variant: NxBannerVariant.success,
              ),
            ),
          Expanded(
            child: AsyncValueWidget(
              value: asyncItems,
              data: (items) {
                if (items.isEmpty) {
                  return NxEmptyState(
                    title: l10n.emptyTitle,
                    body: l10n.emptyMessage,
                    primaryActionLabel: l10n.retry,
                    onPrimaryAction: () => ref.invalidate(couponsListProvider),
                  );
                }
                return ListView.separated(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  itemCount: items.length,
                  separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
                  itemBuilder: (context, index) => _CouponCard(coupon: items[index]),
                );
              },
              error: (e, _) => ErrorView(
                message: localizedCustomerError(context, e),
                onRetry: () => ref.invalidate(couponsListProvider),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _CouponCard extends StatelessWidget {
  const _CouponCard({required this.coupon});

  final Coupon coupon;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final discountLabel = coupon.discountType == CouponDiscountType.percent
        ? '${coupon.discountValue}%'
        : Money(minorUnits: coupon.discountValue, currency: coupon.currency).format();

    return NxCard(
      child: ListTile(
        title: Text(coupon.title.isNotEmpty ? coupon.title : coupon.code),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(coupon.code),
            if (coupon.stackable) Text(l10n.stackable),
            if (coupon.minOrderMinor > 0)
              Text('${l10n.minOrder} ${Money(minorUnits: coupon.minOrderMinor, currency: coupon.currency).format()}'),
          ],
        ),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(discountLabel, style: NxTypography.headlineSm),
            if (coupon.stackable) NxChip(label: l10n.stackable),
          ],
        ),
      ),
    );
  }
}
