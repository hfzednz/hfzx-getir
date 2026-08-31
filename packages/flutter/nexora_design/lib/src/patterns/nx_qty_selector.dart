import 'package:flutter/material.dart';

import '../components/nx_icon_button.dart';
import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/typography.dart';

enum NxQtySelectorSize { sm, md }

/// Quantity stepper — minus, tabular value, plus.
class NxQtySelector extends StatelessWidget {
  const NxQtySelector({
    super.key,
    required this.quantity,
    required this.onIncrement,
    required this.onDecrement,
    this.min = 0,
    this.max = 99,
    this.size = NxQtySelectorSize.md,
    this.semanticLabel,
    this.increaseSemanticLabel,
    this.decreaseSemanticLabel,
  });

  final int quantity;
  final VoidCallback onIncrement;
  final VoidCallback onDecrement;
  final int min;
  final int max;
  final NxQtySelectorSize size;
  final String? semanticLabel;
  final String? increaseSemanticLabel;
  final String? decreaseSemanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final buttonSize = size == NxQtySelectorSize.sm
        ? NxIconButtonSize.sm
        : NxIconButtonSize.md;

    return Semantics(
      label: semanticLabel ?? 'Quantity $quantity',
      child: Container(
        decoration: BoxDecoration(
          border: Border.all(color: colors.borderDefault),
          borderRadius: BorderRadius.circular(NxRadius.md),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            NxIconButton(
              icon: const Icon(Icons.remove),
              size: buttonSize,
              onPressed: quantity <= min ? null : onDecrement,
              semanticLabel: decreaseSemanticLabel ?? 'Decrease quantity',
              disabled: quantity <= min,
            ),
            SizedBox(
              width: size == NxQtySelectorSize.sm ? 28 : 36,
              child: Text(
                '$quantity',
                textAlign: TextAlign.center,
                style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
              ),
            ),
            NxIconButton(
              icon: const Icon(Icons.add),
              size: buttonSize,
              onPressed: quantity >= max ? null : onIncrement,
              semanticLabel: increaseSemanticLabel ?? 'Increase quantity',
              disabled: quantity >= max,
            ),
          ],
        ),
      ),
    );
  }
}
