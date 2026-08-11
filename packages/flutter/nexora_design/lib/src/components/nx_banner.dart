import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxBannerVariant { info, warning, danger, success }

/// Inline system banner — dismissible when non-critical.
class NxBanner extends StatelessWidget {
  const NxBanner({
    super.key,
    required this.message,
    this.variant = NxBannerVariant.info,
    this.title,
    this.onDismiss,
    this.actionLabel,
    this.onAction,
    this.semanticLabel,
  });

  final String message;
  final NxBannerVariant variant;
  final String? title;
  final VoidCallback? onDismiss;
  final String? actionLabel;
  final VoidCallback? onAction;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final (bg, fg, icon) = switch (variant) {
      NxBannerVariant.info => (colors.infoSurface, colors.info, Icons.info_outline),
      NxBannerVariant.warning => (colors.warningSurface, colors.warning, Icons.warning_amber_outlined),
      NxBannerVariant.danger => (colors.dangerSurface, colors.danger, Icons.error_outline),
      NxBannerVariant.success => (colors.successSurface, colors.success, Icons.check_circle_outline),
    };

    return Semantics(
      label: semanticLabel ?? message,
      child: Material(
        color: bg,
        borderRadius: BorderRadius.circular(NxRadius.md),
        child: Padding(
          padding: const EdgeInsets.all(NxSpacing.s3),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(icon, color: fg, size: NxIconSize.md),
              const SizedBox(width: NxSpacing.s3),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (title != null)
                      Text(
                        title!,
                        style: NxTypography.titleSm.copyWith(color: fg),
                      ),
                    Text(
                      message,
                      style: NxTypography.bodySm.copyWith(color: colors.textPrimary),
                    ),
                    if (actionLabel != null) ...[
                      const SizedBox(height: NxSpacing.s2),
                      TextButton(
                        onPressed: onAction,
                        style: TextButton.styleFrom(
                          foregroundColor: fg,
                          padding: EdgeInsets.zero,
                          minimumSize: Size.zero,
                          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                        ),
                        child: Text(
                          actionLabel!,
                          style: NxTypography.bodySm.copyWith(
                            color: fg,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              if (onDismiss != null)
                IconButton(
                  icon: Icon(Icons.close, color: colors.iconSecondary, size: NxIconSize.sm),
                  onPressed: onDismiss,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                ),
            ],
          ),
        ),
      ),
    );
  }
}
