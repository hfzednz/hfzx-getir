import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/elevation.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_button.dart';

/// Modal dialog — max width 400, scrim overlay, stacked actions on mobile.
class NxDialog extends StatelessWidget {
  const NxDialog({
    super.key,
    required this.title,
    this.body,
    this.primaryActionLabel,
    this.secondaryActionLabel,
    this.onPrimary,
    this.onSecondary,
    this.destructive = false,
  });

  final String title;
  final Widget? body;
  final String? primaryActionLabel;
  final String? secondaryActionLabel;
  final VoidCallback? onPrimary;
  final VoidCallback? onSecondary;
  final bool destructive;

  static Future<T?> show<T>({
    required BuildContext context,
    required String title,
    Widget? body,
    String? primaryActionLabel,
    String? secondaryActionLabel,
    VoidCallback? onPrimary,
    VoidCallback? onSecondary,
    bool destructive = false,
  }) {
    return showDialog<T>(
      context: context,
      barrierColor: context.nxColors.bgOverlay,
      builder: (context) => NxDialog(
        title: title,
        body: body,
        primaryActionLabel: primaryActionLabel,
        secondaryActionLabel: secondaryActionLabel,
        onPrimary: onPrimary,
        onSecondary: onSecondary,
        destructive: destructive,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;
    final isMobile = MediaQuery.sizeOf(context).width < 600;

    return Dialog(
      backgroundColor: colors.bgSurface,
      elevation: 0,
      insetPadding: const EdgeInsets.all(NxSpacing.s4),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(NxRadius.lg),
      ),
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 400),
        child: DecoratedBox(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(NxRadius.lg),
            boxShadow: NxElevation.forBrightness(brightness, 3),
          ),
          child: Padding(
            padding: const EdgeInsets.all(NxSpacing.s6),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  title,
                  style: NxTypography.headlineSm.copyWith(color: colors.textPrimary),
                ),
                if (body != null) ...[
                  const SizedBox(height: NxSpacing.s3),
                  DefaultTextStyle(
                    style: NxTypography.bodyMd.copyWith(color: colors.textSecondary),
                    child: body!,
                  ),
                ],
                const SizedBox(height: NxSpacing.s6),
                if (isMobile)
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: _actions(context),
                  )
                else
                  Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: _actions(context),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  List<Widget> _actions(BuildContext context) {
    final actions = <Widget>[];
    if (secondaryActionLabel != null) {
      actions.add(
        NxButton(
          label: secondaryActionLabel!,
          variant: NxButtonVariant.secondary,
          onPressed: () {
            Navigator.of(context).pop();
            onSecondary?.call();
          },
        ),
      );
    }
    if (primaryActionLabel != null) {
      if (actions.isNotEmpty) {
        actions.add(const SizedBox(width: NxSpacing.s2, height: NxSpacing.s2));
      }
      actions.add(
        NxButton(
          label: primaryActionLabel!,
          variant: destructive ? NxButtonVariant.destructive : NxButtonVariant.primary,
          onPressed: () {
            Navigator.of(context).pop();
            onPrimary?.call();
          },
        ),
      );
    }
    return actions;
  }
}
