import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Filter chip — selected state uses brand fill.
class NxChip extends StatelessWidget {
  const NxChip({
    super.key,
    required this.label,
    this.selected = false,
    this.onSelected,
    this.leading,
    this.semanticLabel,
  });

  final String label;
  final bool selected;
  final ValueChanged<bool>? onSelected;
  final Widget? leading;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final bg = selected ? colors.bgBrand : colors.bgSurface;
    final fg = selected ? colors.textOnBrand : colors.textPrimary;
    final border = selected ? colors.bgBrand : colors.borderDefault;

    return Semantics(
      button: true,
      selected: selected,
      label: semanticLabel ?? label,
      child: Material(
        color: bg,
        borderRadius: BorderRadius.circular(NxRadius.md),
        child: InkWell(
          onTap: onSelected == null ? null : () => onSelected!(!selected),
          borderRadius: BorderRadius.circular(NxRadius.md),
          child: Ink(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(NxRadius.md),
              border: Border.all(color: border),
            ),
            child: Padding(
              padding: const EdgeInsets.symmetric(
                horizontal: NxSpacing.s3,
                vertical: NxSpacing.s2,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (leading != null) ...[
                    IconTheme(
                      data: IconThemeData(color: fg, size: NxIconSize.sm),
                      child: leading!,
                    ),
                    const SizedBox(width: NxSpacing.s1),
                  ],
                  Text(
                    label,
                    style: NxTypography.bodySm.copyWith(color: fg),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
