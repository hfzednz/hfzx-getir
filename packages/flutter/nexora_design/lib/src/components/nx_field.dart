import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Text field with label outside, helper/error below.
class NxField extends StatefulWidget {
  const NxField({
    super.key,
    this.label,
    this.hint,
    this.helper,
    this.error,
    this.controller,
    this.onChanged,
    this.obscureText = false,
    this.readOnly = false,
    this.enabled = true,
    this.keyboardType,
    this.textInputAction,
    this.leading,
    this.trailing,
    this.semanticLabel,
    this.autofocus = false,
    this.maxLines = 1,
  });

  final String? label;
  final String? hint;
  final String? helper;
  final String? error;
  final TextEditingController? controller;
  final ValueChanged<String>? onChanged;
  final bool obscureText;
  final bool readOnly;
  final bool enabled;
  final TextInputType? keyboardType;
  final TextInputAction? textInputAction;
  final Widget? leading;
  final Widget? trailing;
  final String? semanticLabel;
  final bool autofocus;
  final int maxLines;

  @override
  State<NxField> createState() => _NxFieldState();
}

class _NxFieldState extends State<NxField> {
  late final FocusNode _focusNode;
  bool _focused = false;

  @override
  void initState() {
    super.initState();
    _focusNode = FocusNode()..addListener(_handleFocus);
  }

  void _handleFocus() => setState(() => _focused = _focusNode.hasFocus);

  @override
  void dispose() {
    _focusNode
      ..removeListener(_handleFocus)
      ..dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final density = context.nx;
    final minHeight = density.density == NxDensity.dense ? 40.0 : 48.0;
    final hasError = widget.error != null && widget.error!.isNotEmpty;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (widget.label != null) ...[
          Text(
            widget.label!,
            style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
          ),
          const SizedBox(height: NxSpacing.s1),
        ],
        Semantics(
          textField: true,
          label: widget.semanticLabel ?? widget.label,
          child: AnimatedContainer(
            duration: const Duration(milliseconds: 120),
            constraints: BoxConstraints(minHeight: minHeight),
            decoration: BoxDecoration(
              color: widget.enabled ? colors.bgSurface : colors.bgDisabled,
              borderRadius: BorderRadius.circular(NxRadius.md),
              border: Border.all(
                color: hasError
                    ? colors.borderDanger
                    : _focused
                        ? colors.borderFocus
                        : colors.borderDefault,
                width: hasError || _focused ? NxBorderWidth.thick : NxBorderWidth.hairline,
              ),
            ),
            child: Row(
              children: [
                if (widget.leading != null) ...[
                  Padding(
                    padding: const EdgeInsets.only(left: NxSpacing.s3),
                    child: IconTheme(
                      data: IconThemeData(color: colors.iconSecondary, size: NxIconSize.md),
                      child: widget.leading!,
                    ),
                  ),
                ],
                Expanded(
                  child: TextField(
                    controller: widget.controller,
                    focusNode: _focusNode,
                    onChanged: widget.onChanged,
                    obscureText: widget.obscureText,
                    readOnly: widget.readOnly,
                    enabled: widget.enabled,
                    keyboardType: widget.keyboardType,
                    textInputAction: widget.textInputAction,
                    autofocus: widget.autofocus,
                    maxLines: widget.maxLines,
                    style: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
                    decoration: InputDecoration(
                      hintText: widget.hint,
                      border: InputBorder.none,
                      enabledBorder: InputBorder.none,
                      focusedBorder: InputBorder.none,
                      errorBorder: InputBorder.none,
                      contentPadding: EdgeInsets.symmetric(
                        horizontal: widget.leading == null ? NxSpacing.s4 : NxSpacing.s2,
                        vertical: density.scaledSpacing(NxSpacing.s3),
                      ),
                      isDense: true,
                    ),
                  ),
                ),
                if (widget.trailing != null)
                  Padding(
                    padding: const EdgeInsets.only(right: NxSpacing.s3),
                    child: widget.trailing!,
                  ),
              ],
            ),
          ),
        ),
        if (hasError) ...[
          const SizedBox(height: NxSpacing.s1),
          Text(
            widget.error!,
            style: NxTypography.captionMd.copyWith(color: colors.danger),
          ),
        ] else if (widget.helper != null) ...[
          const SizedBox(height: NxSpacing.s1),
          Text(
            widget.helper!,
            style: NxTypography.captionMd.copyWith(color: colors.textTertiary),
          ),
        ],
      ],
    );
  }
}
