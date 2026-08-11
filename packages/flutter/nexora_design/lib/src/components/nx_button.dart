import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/colors.dart';
import '../tokens/motion.dart';
import '../tokens/opacity.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_spinner.dart';

enum NxButtonVariant {
  primary,
  accent,
  secondary,
  tertiary,
  destructive,
  inverse,
}

enum NxButtonSize { sm, md, lg }

/// NEXORA button — presentational only, token-driven styling.
class NxButton extends StatefulWidget {
  const NxButton({
    super.key,
    required this.label,
    this.variant = NxButtonVariant.primary,
    this.size = NxButtonSize.md,
    this.leadingIcon,
    this.trailingIcon,
    this.loading = false,
    this.disabled = false,
    this.onPressed,
    this.semanticLabel,
    this.expand = false,
  });

  final String label;
  final NxButtonVariant variant;
  final NxButtonSize size;
  final Widget? leadingIcon;
  final Widget? trailingIcon;
  final bool loading;
  final bool disabled;
  final VoidCallback? onPressed;
  final String? semanticLabel;
  final bool expand;

  @override
  State<NxButton> createState() => _NxButtonState();
}

class _NxButtonState extends State<NxButton> {
  bool _pressed = false;

  bool get _isDisabled => widget.disabled || widget.loading;

  double get _minHeight => switch (widget.size) {
        NxButtonSize.sm => 40,
        NxButtonSize.md => 48,
        NxButtonSize.lg => 56,
      };

  EdgeInsets get _padding => switch (widget.size) {
        NxButtonSize.sm => const EdgeInsets.symmetric(
            horizontal: NxSpacing.s4,
            vertical: NxSpacing.s2,
          ),
        NxButtonSize.md => const EdgeInsets.symmetric(
            horizontal: NxSpacing.s5,
            vertical: NxSpacing.s3,
          ),
        NxButtonSize.lg => const EdgeInsets.symmetric(
            horizontal: NxSpacing.s6,
            vertical: NxSpacing.s4,
          ),
      };

  TextStyle get _labelStyle => switch (widget.size) {
        NxButtonSize.sm => NxTypography.buttonSm,
        NxButtonSize.md => NxTypography.buttonMd,
        NxButtonSize.lg => NxTypography.buttonLg,
      };

  double get _spinnerSize => switch (widget.size) {
        NxButtonSize.sm => 16,
        NxButtonSize.md => 20,
        NxButtonSize.lg => 24,
      };

  _ButtonColors _resolveColors(NxColorRoles colors) {
    switch (widget.variant) {
      case NxButtonVariant.primary:
        return _ButtonColors(
          background: colors.bgBrand,
          foreground: colors.textOnBrand,
          border: Colors.transparent,
        );
      case NxButtonVariant.accent:
        return _ButtonColors(
          background: colors.bgAccent,
          foreground: colors.textOnAccent,
          border: Colors.transparent,
        );
      case NxButtonVariant.secondary:
        return _ButtonColors(
          background: colors.bgSurface,
          foreground: colors.textPrimary,
          border: colors.borderDefault,
        );
      case NxButtonVariant.tertiary:
        return _ButtonColors(
          background: Colors.transparent,
          foreground: colors.textBrand,
          border: Colors.transparent,
        );
      case NxButtonVariant.destructive:
        return _ButtonColors(
          background: colors.danger,
          foreground: colors.textOnBrand,
          border: Colors.transparent,
        );
      case NxButtonVariant.inverse:
        return _ButtonColors(
          background: colors.textInverse.withValues(alpha: 0.12),
          foreground: colors.textInverse,
          border: Colors.transparent,
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final resolved = _resolveColors(colors);
    final bg = _isDisabled
        ? colors.bgDisabled
        : _pressed
            ? Color.alphaBlend(
                colors.textPrimary.withValues(alpha: NxOpacity.pressed),
                resolved.background,
              )
            : resolved.background;
    final fg = _isDisabled ? colors.textDisabled : resolved.foreground;

    final child = Row(
      mainAxisSize: widget.expand ? MainAxisSize.max : MainAxisSize.min,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        if (widget.loading)
          NxSpinner(dimension: _spinnerSize, color: fg)
        else ...[
          if (widget.leadingIcon != null) ...[
            IconTheme(
              data: IconThemeData(color: fg, size: NxIconSize.sm),
              child: widget.leadingIcon!,
            ),
            const SizedBox(width: NxSpacing.s2),
          ],
          Flexible(
            child: Text(
              widget.label,
              style: _labelStyle.copyWith(color: fg),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
            ),
          ),
          if (widget.trailingIcon != null) ...[
            const SizedBox(width: NxSpacing.s2),
            IconTheme(
              data: IconThemeData(color: fg, size: NxIconSize.sm),
              child: widget.trailingIcon!,
            ),
          ],
        ],
      ],
    );

    return Semantics(
      button: true,
      enabled: !_isDisabled,
      label: widget.semanticLabel ?? widget.label,
      child: AnimatedScale(
        scale: _pressed && !_isDisabled ? 0.98 : 1.0,
        duration: NxDuration.fast,
        curve: NxCurves.standard,
        child: Material(
          color: bg,
          elevation: 0,
          borderRadius: BorderRadius.circular(NxRadius.md),
          child: InkWell(
            onTap: _isDisabled ? null : widget.onPressed,
            onHighlightChanged: (v) => setState(() => _pressed = v),
            borderRadius: BorderRadius.circular(NxRadius.md),
            splashColor: colors.bgBrand.withValues(alpha: NxOpacity.hover),
            child: Ink(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(NxRadius.md),
                border: resolved.border == Colors.transparent
                    ? null
                    : Border.all(color: resolved.border),
              ),
              child: ConstrainedBox(
                constraints: BoxConstraints(minHeight: _minHeight),
                child: Padding(
                  padding: _padding,
                  child: widget.expand ? Center(child: child) : child,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ButtonColors {
  const _ButtonColors({
    required this.background,
    required this.foreground,
    required this.border,
  });

  final Color background;
  final Color foreground;
  final Color border;
}
