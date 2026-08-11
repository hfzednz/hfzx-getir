import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxTagVariant { neutral, brand, accent, danger, warning }

/// Static meta tag — promo, cold chain, fragile, etc.
class NxTag extends StatelessWidget {
  const NxTag({
    super.key,
    required this.label,
    this.variant = NxTagVariant.neutral,
    this.icon,
    this.semanticLabel,
  });

  final String label;
  final NxTagVariant variant;
  final IconData? icon;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final (bg, fg) = switch (variant) {
      NxTagVariant.neutral => (colors.bgSunken, colors.textSecondary),
      NxTagVariant.brand => (colors.infoSurface, colors.textBrand),
      NxTagVariant.accent => (colors.bgAccent.withValues(alpha: 0.35), colors.textOnAccent),
      NxTagVariant.danger => (colors.dangerSurface, colors.danger),
      NxTagVariant.warning => (colors.warningSurface, colors.warning),
    };

    return Semantics(
      label: semanticLabel ?? label,
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: NxSpacing.s2,
          vertical: NxSpacing.s1,
        ),
        decoration: BoxDecoration(
          color: bg,
          borderRadius: BorderRadius.circular(NxRadius.xs),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (icon != null) ...[
              Icon(icon, size: NxIconSize.xs, color: fg),
              const SizedBox(width: NxSpacing.s1),
            ],
            Text(
              label,
              style: NxTypography.captionSm.copyWith(color: fg),
            ),
          ],
        ),
      ),
    );
  }
}
