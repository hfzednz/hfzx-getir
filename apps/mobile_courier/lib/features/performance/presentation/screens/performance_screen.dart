import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/performance_entity.dart';
import '../providers/performance_providers.dart';

class PerformanceScreen extends ConsumerWidget {
  const PerformanceScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(performanceProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Performance'),
      body: AsyncValueWidget<PerformanceMetrics>(
        value: async,
        data: (m) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s4),
          children: [
            _metric('Acceptance', m.acceptanceRate, colors),
            _metric('Completion', m.completionRate, colors),
            _metric('On-time', m.onTimeRate, colors),
            _metric('Rating', m.rating, colors, asPercent: false),
            _metric('Safety score', m.safetyScore, colors, asPercent: false),
          ],
        ),
      ),
    );
  }

  Widget _metric(
    String label,
    double value,
    NxColorRoles colors, {
    bool asPercent = true,
  }) {
    final display = asPercent
        ? '${(value * 100).toStringAsFixed(0)}%'
        : value.toStringAsFixed(1);
    return Padding(
      padding: const EdgeInsets.only(bottom: NxSpacing.s2),
      child: NxCard(
        child: Row(
          children: [
            Expanded(
              child: Text(
                label,
                style: NxTypography.bodyMd.copyWith(color: colors.textSecondary),
              ),
            ),
            Text(
              display,
              style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
            ),
          ],
        ),
      ),
    );
  }
}
