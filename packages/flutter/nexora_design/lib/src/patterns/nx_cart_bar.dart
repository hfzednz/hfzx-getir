import 'package:flutter/material.dart';

import '../components/nx_button.dart';
import '../theme/nx_theme.dart';
import '../tokens/elevation.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_price_block.dart';

/// Sticky cart affordance — item count, total, view cart CTA.
class NxCartBar extends StatelessWidget {
  const NxCartBar({
    super.key,
    required this.itemCount,
    required this.total,
    required this.onViewCart,
    this.currencySymbol = '₺',
    this.ctaLabel = 'View cart',
    this.semanticLabel,
  });

  final int itemCount;
  final String total;
  final VoidCallback onViewCart;
  final String currencySymbol;
  final String ctaLabel;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    if (itemCount <= 0) return const SizedBox.shrink();

    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;

    return Semantics(
      label: semanticLabel ?? '$itemCount items, total $currencySymbol$total',
      button: true,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.bgSurface,
          boxShadow: NxElevation.forBrightness(brightness, 2),
          border: Border(top: BorderSide(color: colors.borderSubtle)),
        ),
        child: SafeArea(
          top: false,
          child: Padding(
            padding: const EdgeInsets.symmetric(
              horizontal: NxSpacing.s4,
              vertical: NxSpacing.s3,
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: NxSpacing.s3,
                    vertical: NxSpacing.s2,
                  ),
                  decoration: BoxDecoration(
                    color: colors.bgBrand,
                    borderRadius: BorderRadius.circular(NxRadius.md),
                  ),
                  child: Text(
                    '$itemCount',
                    style: NxTypography.titleSm.copyWith(color: colors.textOnBrand),
                  ),
                ),
                const SizedBox(width: NxSpacing.s3),
                Expanded(
                  child: NxPriceBlock(
                    price: total,
                    currencySymbol: currencySymbol,
                    size: NxPriceBlockSize.md,
                  ),
                ),
                NxButton(
                  label: ctaLabel,
                  variant: NxButtonVariant.accent,
                  onPressed: onViewCart,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
