import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/elevation.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';

enum NxSheetHeight { content, mid, full }

/// Bottom sheet helpers — grabber, rounded top, drag dismiss.
abstract final class NxSheet {
  static Future<T?> show<T>({
    required BuildContext context,
    required Widget child,
    NxSheetHeight height = NxSheetHeight.content,
    bool isDismissible = true,
    bool enableDrag = true,
  }) {
    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;

    return showModalBottomSheet<T>(
      context: context,
      isScrollControlled: true,
      isDismissible: isDismissible,
      enableDrag: enableDrag,
      backgroundColor: Colors.transparent,
      barrierColor: colors.bgOverlay,
      transitionAnimationController: null,
      builder: (context) {
        final media = MediaQuery.of(context);
        final maxHeight = switch (height) {
          NxSheetHeight.content => media.size.height * 0.45,
          NxSheetHeight.mid => media.size.height * 0.65,
          NxSheetHeight.full => media.size.height * 0.92,
        };

        return Padding(
          padding: EdgeInsets.only(bottom: media.viewInsets.bottom),
          child: ConstrainedBox(
            constraints: BoxConstraints(maxHeight: maxHeight),
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: colors.bgSurface,
                borderRadius: const BorderRadius.vertical(
                  top: Radius.circular(NxRadius.lg),
                ),
                boxShadow: NxElevation.forBrightness(brightness, 3),
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const SizedBox(height: NxSpacing.s2),
                  Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: colors.borderDefault,
                      borderRadius: BorderRadius.circular(NxRadius.full),
                    ),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  Flexible(child: child),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}
