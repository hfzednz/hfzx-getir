import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/report_kpis.dart';
import '../providers/reports_providers.dart';

class ReportsScreen extends ConsumerWidget {
  const ReportsScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(reportKpisProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Reports'),
      body: AsyncValueWidget<ReportKpis>(
        value: async,
        data: (k) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s3),
          children: [
            _KpiCard(title: 'Pick speed', value: '${k.pickSpeedPerHour.toStringAsFixed(0)} /h'),
            _KpiCard(title: 'Pick accuracy', value: '${(k.pickAccuracy * 100).toStringAsFixed(1)}%'),
            _KpiCard(title: 'Pack speed', value: '${k.packSpeedPerHour.toStringAsFixed(0)} /h'),
            _KpiCard(title: 'Waste', value: '${k.wasteUnits} units'),
          ],
        ),
      ),
    );
  }
}

class _KpiCard extends StatelessWidget {
  const _KpiCard({required this.title, required this.value});
  final String title;
  final String value;
  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: NxSpacing.s2),
      child: NxCard(
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(title, style: NxTypography.bodyMd),
            Text(value, style: NxTypography.titleMd),
          ],
        ),
      ),
    );
  }
}
