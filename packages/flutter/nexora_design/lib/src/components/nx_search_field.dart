import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/spacing.dart';
import 'nx_field.dart';

/// Search input with leading magnifier and optional clear action.
class NxSearchField extends StatefulWidget {
  const NxSearchField({
    super.key,
    this.hint = 'Search',
    this.controller,
    this.onChanged,
    this.onSubmitted,
    this.onClear,
    this.autofocus = false,
    this.semanticLabel,
  });

  final String hint;
  final TextEditingController? controller;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onSubmitted;
  final VoidCallback? onClear;
  final bool autofocus;
  final String? semanticLabel;

  @override
  State<NxSearchField> createState() => _NxSearchFieldState();
}

class _NxSearchFieldState extends State<NxSearchField> {
  late TextEditingController _controller;
  bool _ownsController = false;

  @override
  void initState() {
    super.initState();
    if (widget.controller != null) {
      _controller = widget.controller!;
    } else {
      _controller = TextEditingController();
      _ownsController = true;
    }
    _controller.addListener(_onTextChanged);
  }

  void _onTextChanged() => setState(() {});

  @override
  void dispose() {
    _controller.removeListener(_onTextChanged);
    if (_ownsController) _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final hasText = _controller.text.isNotEmpty;

    return NxField(
      controller: _controller,
      hint: widget.hint,
      autofocus: widget.autofocus,
      semanticLabel: widget.semanticLabel ?? 'Search field',
      onChanged: widget.onChanged,
      textInputAction: TextInputAction.search,
      leading: Icon(Icons.search, color: colors.iconSecondary),
      trailing: hasText
          ? IconButton(
              icon: Icon(Icons.close, color: colors.iconSecondary, size: NxIconSize.sm),
              onPressed: () {
                _controller.clear();
                widget.onClear?.call();
                widget.onChanged?.call('');
              },
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
            )
          : null,
    );
  }
}
