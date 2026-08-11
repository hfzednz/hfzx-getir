import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/elevation.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';

enum NxCardVariant { interactive, staticCard, outlined, elevated }

/// Surface container — border OR elevation, not both heavy.
class NxCard extends StatelessWidget {
  const NxCard({
    super.key,
    required this.child,
    this.variant = NxCardVariant.outlined,
    this.onTap,
    this.padding,
    this.semanticLabel,
  });

  final Widget child;
  final NxCardVariant variant;
  final VoidCallback? onTap;
  final EdgeInsetsGeometry? padding;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;
    final density = context.nx;
    final resolvedPadding = padding ??
        EdgeInsets.all(
          density.density == NxDensity.dense ? NxSpacing.s3 : NxSpacing.s4,
        );

    final useElevation = variant == NxCardVariant.elevated;
    final useBorder = variant == NxCardVariant.outlined ||
        variant == NxCardVariant.interactive ||
        (variant == NxCardVariant.staticCard && brightness == Brightness.dark);

    final decoration = BoxDecoration(
      color: colors.bgSurface,
      borderRadius: BorderRadius.circular(NxRadius.md),
      border: useBorder && !useElevation
          ? Border.all(color: colors.borderSubtle)
          : null,
      boxShadow: useElevation ? NxElevation.forBrightness(brightness, 1) : null,
    );

    Widget content = Padding(padding: resolvedPadding, child: child);

    if (onTap != null || variant == NxCardVariant.interactive) {
      content = Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(NxRadius.md),
          child: content,
        ),
      );
    }

    return Semantics(
      container: true,
      button: onTap != null,
      label: semanticLabel,
      child: DecoratedBox(
        decoration: decoration,
        child: content,
      ),
    );
  }
}
