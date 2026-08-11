import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Commerce discount badge — danger wash for %-off.
class NxDiscountBadge extends StatelessWidget {
  const NxDiscountBadge({
    super.key,
    required this.label,
    this.semanticLabel,
  });

  final String label;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? label,
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: NxSpacing.s2,
          vertical: NxSpacing.s1,
        ),
        decoration: BoxDecoration(
          color: colors.dangerSurface,
          borderRadius: BorderRadius.circular(NxRadius.xs),
        ),
        child: Text(
          label,
          style: NxTypography.captionSm.copyWith(color: colors.danger),
        ),
      ),
    );
  }
}
