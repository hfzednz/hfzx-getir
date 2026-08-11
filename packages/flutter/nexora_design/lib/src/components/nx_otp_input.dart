import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// OTP input — 6 boxes with auto-advance and paste support.
class NxOtpInput extends StatefulWidget {
  const NxOtpInput({
    super.key,
    this.length = 6,
    this.onCompleted,
    this.onChanged,
    this.semanticLabel,
    this.enabled = true,
  });

  final int length;
  final ValueChanged<String>? onCompleted;
  final ValueChanged<String>? onChanged;
  final String? semanticLabel;
  final bool enabled;

  @override
  State<NxOtpInput> createState() => _NxOtpInputState();
}

class _NxOtpInputState extends State<NxOtpInput> {
  late final List<TextEditingController> _controllers;
  late final List<FocusNode> _focusNodes;

  @override
  void initState() {
    super.initState();
    _controllers = List.generate(widget.length, (_) => TextEditingController());
    _focusNodes = List.generate(widget.length, (_) => FocusNode());
  }

  @override
  void dispose() {
    for (final c in _controllers) {
      c.dispose();
    }
    for (final f in _focusNodes) {
      f.dispose();
    }
    super.dispose();
  }

  String get _value => _controllers.map((c) => c.text).join();

  void _notify() {
    final value = _value;
    widget.onChanged?.call(value);
    if (value.length == widget.length) {
      widget.onCompleted?.call(value);
    }
  }

  void _handleChanged(int index, String value) {
    if (value.length > 1) {
      _handlePaste(value);
      return;
    }

    if (value.isNotEmpty && index < widget.length - 1) {
      _focusNodes[index + 1].requestFocus();
    }
    if (value.isEmpty && index > 0) {
      _focusNodes[index - 1].requestFocus();
    }
    _notify();
  }

  void _handlePaste(String raw) {
    final digits = raw.replaceAll(RegExp(r'\D'), '');
    for (var i = 0; i < widget.length; i++) {
      _controllers[i].text = i < digits.length ? digits[i] : '';
    }
    if (digits.length >= widget.length) {
      _focusNodes.last.requestFocus();
    } else if (digits.isNotEmpty) {
      _focusNodes[digits.length.clamp(0, widget.length - 1)].requestFocus();
    }
    _notify();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: widget.semanticLabel ?? 'One-time password input',
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: List.generate(widget.length, (index) {
          return Expanded(
            child: Padding(
              padding: EdgeInsets.only(
                right: index < widget.length - 1 ? NxSpacing.s2 : 0,
              ),
              child: ListenableBuilder(
                listenable: _focusNodes[index],
                builder: (context, _) {
                  final focused = _focusNodes[index].hasFocus;
                  return AnimatedContainer(
                    duration: const Duration(milliseconds: 120),
                    height: 48,
                    decoration: BoxDecoration(
                      color: widget.enabled ? colors.bgSurface : colors.bgDisabled,
                      borderRadius: BorderRadius.circular(NxRadius.md),
                      border: Border.all(
                        color: focused ? colors.borderFocus : colors.borderDefault,
                        width: focused ? NxBorderWidth.thick : NxBorderWidth.hairline,
                      ),
                    ),
                    child: TextField(
                      controller: _controllers[index],
                      focusNode: _focusNodes[index],
                      enabled: widget.enabled,
                      textAlign: TextAlign.center,
                      keyboardType: TextInputType.number,
                      maxLength: 1,
                      style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
                      inputFormatters: [FilteringTextInputFormatter.digitsOnly],
                      decoration: const InputDecoration(
                        counterText: '',
                        border: InputBorder.none,
                        contentPadding: EdgeInsets.zero,
                      ),
                      onChanged: (v) => _handleChanged(index, v),
                    ),
                  );
                },
              ),
            ),
          );
        }),
      ),
    );
  }
}
