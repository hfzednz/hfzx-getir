import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Price display with tabular numerals.
class NxPriceBlock extends StatelessWidget {
  const NxPriceBlock({
    super.key,
    required this.price,
    this.originalPrice,
    this.perUnit,
    this.currencySymbol = '₺',
    this.size = NxPriceBlockSize.md,
    this.semanticLabel,
  });

  final String price;
  final String? originalPrice;
  final String? perUnit;
  final String currencySymbol;
  final NxPriceBlockSize size;
  final String? semanticLabel;

  TextStyle get _priceStyle => switch (size) {
        NxPriceBlockSize.lg => NxTypography.priceLg,
        NxPriceBlockSize.md => NxTypography.priceMd,
        NxPriceBlockSize.sm => NxTypography.priceSm,
      };

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? 'Price $currencySymbol$price',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.baseline,
            textBaseline: TextBaseline.alphabetic,
            children: [
              Text(
                '$currencySymbol$price',
                style: _priceStyle.copyWith(color: colors.textPrimary),
              ),
              if (originalPrice != null) ...[
                const SizedBox(width: NxSpacing.s2),
                Text(
                  '$currencySymbol$originalPrice',
                  style: NxTypography.bodySm.copyWith(
                    color: colors.textTertiary,
                    decoration: TextDecoration.lineThrough,
                    decorationColor: colors.textTertiary,
                  ),
                ),
              ],
            ],
          ),
          if (perUnit != null)
            Text(
              perUnit!,
              style: NxTypography.captionMd.copyWith(color: colors.textTertiary),
            ),
        ],
      ),
    );
  }
}

enum NxPriceBlockSize { sm, md, lg }
