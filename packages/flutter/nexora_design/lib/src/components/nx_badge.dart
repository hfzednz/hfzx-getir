import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxBadgeVariant { brand, danger, neutral }

/// Count badge for navigation — caps at 99+.
class NxBadge extends StatelessWidget {
  const NxBadge({
    super.key,
    required this.count,
    this.variant = NxBadgeVariant.brand,
    this.semanticLabel,
  });

  final int count;
  final NxBadgeVariant variant;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    if (count <= 0) return const SizedBox.shrink();

    final colors = context.nxColors;
    final (bg, fg) = switch (variant) {
      NxBadgeVariant.brand => (colors.bgBrand, colors.textOnBrand),
      NxBadgeVariant.danger => (colors.danger, colors.textOnBrand),
      NxBadgeVariant.neutral => (colors.bgSunken, colors.textPrimary),
    };

    final label = count > 99 ? '99+' : '$count';

    return Semantics(
      label: semanticLabel ?? '$count notifications',
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: NxSpacing.s2,
          vertical: NxSpacing.s0_5,
        ),
        constraints: const BoxConstraints(minWidth: 20, minHeight: 20),
        decoration: BoxDecoration(
          color: bg,
          borderRadius: BorderRadius.circular(NxRadius.full),
        ),
        alignment: Alignment.center,
        child: Text(
          label,
          style: NxTypography.captionSm.copyWith(color: fg),
        ),
      ),
    );
  }
}
