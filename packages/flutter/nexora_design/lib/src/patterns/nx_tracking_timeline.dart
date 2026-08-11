import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/colors.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

enum NxTrackingStepState { completed, current, upcoming, failed }

class NxTrackingStep {
  const NxTrackingStep({
    required this.label,
    required this.state,
    this.subtitle,
  });

  final String label;
  final NxTrackingStepState state;
  final String? subtitle;
}

/// Vertical order tracking timeline.
class NxTrackingTimeline extends StatelessWidget {
  const NxTrackingTimeline({
    super.key,
    required this.steps,
    this.semanticLabel,
  });

  final List<NxTrackingStep> steps;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Semantics(
      label: semanticLabel ?? 'Order tracking timeline',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: List.generate(steps.length, (index) {
          final step = steps[index];
          final isLast = index == steps.length - 1;
          final (dotColor, lineColor, textColor) = _styleFor(step.state, colors);

          return IntrinsicHeight(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Column(
                  children: [
                    Container(
                      width: 12,
                      height: 12,
                      decoration: BoxDecoration(
                        color: dotColor,
                        shape: BoxShape.circle,
                        border: step.state == NxTrackingStepState.current
                            ? Border.all(color: colors.bgBrand, width: 2)
                            : null,
                      ),
                    ),
                    if (!isLast)
                      Expanded(
                        child: Container(
                          width: 2,
                          color: lineColor,
                        ),
                      ),
                  ],
                ),
                const SizedBox(width: NxSpacing.s3),
                Expanded(
                  child: Padding(
                    padding: EdgeInsets.only(bottom: isLast ? 0 : NxSpacing.s4),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          step.label,
                          style: NxTypography.titleSm.copyWith(color: textColor),
                        ),
                        if (step.subtitle != null) ...[
                          const SizedBox(height: NxSpacing.s1),
                          Text(
                            step.subtitle!,
                            style: NxTypography.captionMd.copyWith(
                              color: colors.textTertiary,
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
              ],
            ),
          );
        }),
      ),
    );
  }

  (Color, Color, Color) _styleFor(NxTrackingStepState state, NxColorRoles colors) {
    return switch (state) {
      NxTrackingStepState.completed => (
          colors.success,
          colors.borderDefault,
          colors.textPrimary,
        ),
      NxTrackingStepState.current => (
          colors.bgBrand,
          colors.borderSubtle,
          colors.textBrand,
        ),
      NxTrackingStepState.upcoming => (
          colors.borderDefault,
          colors.borderSubtle,
          colors.textTertiary,
        ),
      NxTrackingStepState.failed => (
          colors.danger,
          colors.borderSubtle,
          colors.danger,
        ),
    };
  }
}
