import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/motion.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';

enum NxIconButtonVariant { standard, filled, tonal, danger }

enum NxIconButtonSize { sm, md }

/// Icon-only button with accessible target sizes.
class NxIconButton extends StatefulWidget {
  const NxIconButton({
    super.key,
    required this.icon,
    this.onPressed,
    this.variant = NxIconButtonVariant.standard,
    this.size = NxIconButtonSize.md,
    this.tooltip,
    this.semanticLabel,
    this.disabled = false,
  });

  final Widget icon;
  final VoidCallback? onPressed;
  final NxIconButtonVariant variant;
  final NxIconButtonSize size;
  final String? tooltip;
  final String? semanticLabel;
  final bool disabled;

  @override
  State<NxIconButton> createState() => _NxIconButtonState();
}

class _NxIconButtonState extends State<NxIconButton> {
  bool _pressed = false;

  double get _target => widget.size == NxIconButtonSize.sm ? 40 : 48;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final (bg, fg, border) = switch (widget.variant) {
      NxIconButtonVariant.standard => (
          Colors.transparent,
          colors.iconPrimary,
          Colors.transparent,
        ),
      NxIconButtonVariant.filled => (
          colors.bgBrand,
          colors.textOnBrand,
          Colors.transparent,
        ),
      NxIconButtonVariant.tonal => (
          colors.infoSurface,
          colors.iconBrand,
          Colors.transparent,
        ),
      NxIconButtonVariant.danger => (
          colors.dangerSurface,
          colors.danger,
          Colors.transparent,
        ),
    };

    final effectiveBg = widget.disabled ? colors.bgDisabled : bg;
    final effectiveFg = widget.disabled ? colors.textDisabled : fg;

    Widget button = Semantics(
      button: true,
      enabled: !widget.disabled,
      label: widget.semanticLabel ?? widget.tooltip,
      child: AnimatedScale(
        scale: _pressed && !widget.disabled ? 0.98 : 1.0,
        duration: NxDuration.fast,
        curve: NxCurves.standard,
        child: Material(
          color: effectiveBg,
          borderRadius: BorderRadius.circular(NxRadius.md),
          child: InkWell(
            onTap: widget.disabled ? null : widget.onPressed,
            onHighlightChanged: (v) => setState(() => _pressed = v),
            borderRadius: BorderRadius.circular(NxRadius.md),
            child: Ink(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(NxRadius.md),
                border: border == Colors.transparent
                    ? null
                    : Border.all(color: border),
              ),
              child: SizedBox(
                width: _target,
                height: _target,
                child: IconTheme(
                  data: IconThemeData(
                    color: effectiveFg,
                    size: NxIconSize.md,
                  ),
                  child: Center(child: widget.icon),
                ),
              ),
            ),
          ),
        ),
      ),
    );

    if (widget.tooltip != null) {
      button = Tooltip(message: widget.tooltip!, child: button);
    }

    return button;
  }
}
