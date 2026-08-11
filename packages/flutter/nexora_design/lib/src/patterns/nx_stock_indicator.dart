import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxStockStatus { inStock, low, outOfStock }

/// Stock indicator — color + label, never color-only.
class NxStockIndicator extends StatelessWidget {
  const NxStockIndicator({
    super.key,
    required this.status,
    this.lowStockLabel,
    this.semanticLabel,
  });

  final NxStockStatus status;
  final String? lowStockLabel;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final (dotColor, label) = switch (status) {
      NxStockStatus.inStock => (colors.success, 'In stock'),
      NxStockStatus.low => (colors.warning, lowStockLabel ?? 'Low stock'),
      NxStockStatus.outOfStock => (colors.danger, 'Out of stock'),
    };

    return Semantics(
      label: semanticLabel ?? label,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: dotColor, shape: BoxShape.circle),
          ),
          const SizedBox(width: NxSpacing.s2),
          Text(
            label,
            style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
          ),
        ],
      ),
    );
  }
}
