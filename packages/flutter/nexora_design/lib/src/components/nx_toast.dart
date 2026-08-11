import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxToastVariant { info, success, warning, danger }

/// Snackbar / toast helper with semantic left bar.
abstract final class NxToast {
  static void show(
    BuildContext context, {
    required String message,
    NxToastVariant variant = NxToastVariant.info,
    String? actionLabel,
    VoidCallback? onAction,
    Duration duration = const Duration(seconds: 4),
  }) {
    final colors = context.nxColors;
    final accentColor = switch (variant) {
      NxToastVariant.info => colors.info,
      NxToastVariant.success => colors.success,
      NxToastVariant.warning => colors.warning,
      NxToastVariant.danger => colors.danger,
    };

    final messenger = ScaffoldMessenger.of(context);
    messenger.hideCurrentSnackBar();
    messenger.showSnackBar(
      SnackBar(
        duration: duration,
        behavior: SnackBarBehavior.floating,
        margin: const EdgeInsets.fromLTRB(
          NxSpacing.s4,
          0,
          NxSpacing.s4,
          NxSpacing.s16,
        ),
        backgroundColor: colors.bgSurfaceRaised,
        elevation: 0,
        content: Row(
          children: [
            Container(
              width: 4,
              height: 36,
              decoration: BoxDecoration(
                color: accentColor,
                borderRadius: BorderRadius.circular(NxRadius.xs),
              ),
            ),
            const SizedBox(width: NxSpacing.s3),
            Expanded(
              child: Text(
                message,
                style: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
        action: actionLabel != null
            ? SnackBarAction(
                label: actionLabel,
                onPressed: onAction ?? () {},
                textColor: colors.textLink,
              )
            : null,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          side: BorderSide(color: colors.borderSubtle),
        ),
      ),
    );
  }
}
