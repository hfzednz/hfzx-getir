import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

class NxCheckoutLine {
  const NxCheckoutLine({
    required this.label,
    required this.amount,
    this.emphasized = false,
    this.negative = false,
  });

  final String label;
  final String amount;
  final bool emphasized;
  final bool negative;
}

/// Checkout summary line items with emphasized total.
class NxCheckoutSummary extends StatelessWidget {
  const NxCheckoutSummary({
    super.key,
    required this.lines,
    this.currencySymbol = '₺',
    this.semanticLabel,
  });

  final List<NxCheckoutLine> lines;
  final String currencySymbol;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? 'Checkout summary',
      child: Column(
        children: lines.map((line) {
          final prefix = line.negative ? '−$currencySymbol' : currencySymbol;
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: NxSpacing.s1),
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    line.label,
                    style: (line.emphasized ? NxTypography.titleMd : NxTypography.bodyMd)
                        .copyWith(color: colors.textPrimary),
                  ),
                ),
                Text(
                  '$prefix${line.amount}',
                  style: (line.emphasized ? NxTypography.priceMd : NxTypography.tableCell)
                      .copyWith(color: colors.textPrimary),
                ),
              ],
            ),
          );
        }).toList(),
      ),
    );
  }
}
