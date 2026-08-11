import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Inline delivery ETA chip — not a floating hero overlay.
class NxDeliveryBadge extends StatelessWidget {
  const NxDeliveryBadge({
    super.key,
    required this.etaLabel,
    this.semanticLabel,
  });

  final String etaLabel;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? 'Delivery $etaLabel',
      child: Container(
        padding: const EdgeInsets.symmetric(
          horizontal: NxSpacing.s2,
          vertical: NxSpacing.s1,
        ),
        decoration: BoxDecoration(
          color: colors.infoSurface,
          borderRadius: BorderRadius.circular(NxRadius.xs),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.schedule, size: NxIconSize.sm, color: colors.info),
            const SizedBox(width: NxSpacing.s1),
            Text(
              etaLabel,
              style: NxTypography.etaMd.copyWith(
                color: colors.info,
                fontSize: 12,
                height: 16 / 12,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
