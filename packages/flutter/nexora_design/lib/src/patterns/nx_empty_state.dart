import 'package:flutter/material.dart';

import '../components/nx_button.dart';
import '../theme/nx_theme.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Empty state — illustration slot, title, body, and actions.
class NxEmptyState extends StatelessWidget {
  const NxEmptyState({
    super.key,
    required this.title,
    this.body,
    this.illustration,
    this.primaryActionLabel,
    this.secondaryActionLabel,
    this.onPrimaryAction,
    this.onSecondaryAction,
    this.semanticLabel,
  });

  final String title;
  final String? body;
  final Widget? illustration;
  final String? primaryActionLabel;
  final String? secondaryActionLabel;
  final VoidCallback? onPrimaryAction;
  final VoidCallback? onSecondaryAction;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? title,
      child: Padding(
        padding: const EdgeInsets.all(NxSpacing.s6),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          mainAxisSize: MainAxisSize.min,
          children: [
            if (illustration != null) ...[
              illustration!,
              const SizedBox(height: NxSpacing.s6),
            ],
            Text(
              title,
              textAlign: TextAlign.center,
              style: NxTypography.headlineSm.copyWith(color: colors.textPrimary),
            ),
            if (body != null) ...[
              const SizedBox(height: NxSpacing.s2),
              Text(
                body!,
                textAlign: TextAlign.center,
                style: NxTypography.bodyMd.copyWith(color: colors.textSecondary),
              ),
            ],
            if (primaryActionLabel != null || secondaryActionLabel != null) ...[
              const SizedBox(height: NxSpacing.s6),
              if (primaryActionLabel != null)
                NxButton(
                  label: primaryActionLabel!,
                  expand: true,
                  onPressed: onPrimaryAction,
                ),
              if (secondaryActionLabel != null) ...[
                const SizedBox(height: NxSpacing.s3),
                NxButton(
                  label: secondaryActionLabel!,
                  variant: NxButtonVariant.secondary,
                  expand: true,
                  onPressed: onSecondaryAction,
                ),
              ],
            ],
          ],
        ),
      ),
    );
  }
}
