import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/dashboard_entity.dart';
import '../providers/dashboard_providers.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(dashboardProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: NxTopBar(
        title: 'Dashboard',
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () => context.push(RouteNames.notifications),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(dashboardProvider),
        child: AsyncValueWidget<WarehouseDashboard>(
          value: async,
          data: (dash) => ListView(
            padding: const EdgeInsets.all(NxSpacing.s3),
            children: [
              _QueueGrid(dash: dash),
              const SizedBox(height: NxSpacing.s3),
              _AlertRow(dash: dash),
              const SizedBox(height: NxSpacing.s3),
              Text(
                'KPIs',
                style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
              ),
              const SizedBox(height: NxSpacing.s2),
              _KpiRow(kpis: dash.kpis),
              if (dash.aiTip != null && dash.aiTip!.isNotEmpty) ...[
                const SizedBox(height: NxSpacing.s3),
                NxCard(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(Icons.auto_awesome, color: colors.iconBrand, size: 20),
                      const SizedBox(width: NxSpacing.s2),
                      Expanded(
                        child: Text(
                          dash.aiTip!,
                          style: NxTypography.bodySm
                              .copyWith(color: colors.textPrimary),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _QueueGrid extends StatelessWidget {
  const _QueueGrid({required this.dash});
  final WarehouseDashboard dash;

  @override
  Widget build(BuildContext context) {
    final items = [
      ('Orders', dash.ordersWaiting, RouteNames.home, Icons.receipt_long),
      ('Pick', dash.pickQueue, RouteNames.picking, Icons.shopping_basket),
      ('Pack', dash.packQueue, RouteNames.packing, Icons.inventory_2),
      ('Dispatch', dash.dispatchQueue, RouteNames.dispatch, Icons.local_shipping),
      ('Couriers', dash.courierArrivals, RouteNames.dispatch, Icons.person_pin),
    ];
    return Wrap(
      spacing: NxSpacing.s2,
      runSpacing: NxSpacing.s2,
      children: items.map((e) {
        return SizedBox(
          width: (MediaQuery.sizeOf(context).width - NxSpacing.s3 * 2 - NxSpacing.s2) / 2,
          child: NxCard(
            variant: NxCardVariant.interactive,
            onTap: () => context.go(e.$3),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(e.$4, size: 18),
                const SizedBox(height: NxSpacing.s1),
                Text('${e.$2}', style: NxTypography.headlineSm),
                Text(e.$1, style: NxTypography.captionMd),
              ],
            ),
          ),
        );
      }).toList(),
    );
  }
}

class _AlertRow extends StatelessWidget {
  const _AlertRow({required this.dash});
  final WarehouseDashboard dash;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return Row(
      children: [
        Expanded(
          child: _AlertChip(
            label: 'Low',
            count: dash.lowStockAlerts,
            color: colors.warning,
            onTap: () => context.push(RouteNames.inventory),
          ),
        ),
        const SizedBox(width: NxSpacing.s2),
        Expanded(
          child: _AlertChip(
            label: 'OOS',
            count: dash.oosAlerts,
            color: colors.danger,
            onTap: () => context.push(RouteNames.inventory),
          ),
        ),
        const SizedBox(width: NxSpacing.s2),
        Expanded(
          child: _AlertChip(
            label: 'Expiry',
            count: dash.expiryAlerts,
            color: colors.warning,
            onTap: () => context.push(RouteNames.expiry),
          ),
        ),
      ],
    );
  }
}

class _AlertChip extends StatelessWidget {
  const _AlertChip({
    required this.label,
    required this.count,
    required this.color,
    required this.onTap,
  });

  final String label;
  final int count;
  final Color color;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return NxCard(
      variant: NxCardVariant.interactive,
      onTap: onTap,
      padding: const EdgeInsets.all(NxSpacing.s2),
      child: Column(
        children: [
          Text('$count', style: NxTypography.titleMd.copyWith(color: color)),
          Text(label, style: NxTypography.captionMd),
        ],
      ),
    );
  }
}

class _KpiRow extends StatelessWidget {
  const _KpiRow({required this.kpis});
  final WarehouseKpis kpis;

  @override
  Widget build(BuildContext context) {
    return NxCard(
      child: Column(
        children: [
          _kpi('Pick speed', '${kpis.pickSpeedPerHour.toStringAsFixed(0)}/h'),
          _kpi('Accuracy', '${(kpis.pickAccuracy * 100).toStringAsFixed(1)}%'),
          _kpi('Pack speed', '${kpis.packSpeedPerHour.toStringAsFixed(0)}/h'),
          _kpi('Waste', '${kpis.wasteUnits} u'),
          _kpi(
            'On-time dispatch',
            '${(kpis.onTimeDispatchRate * 100).toStringAsFixed(0)}%',
          ),
        ],
      ),
    );
  }

  Widget _kpi(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: NxSpacing.s1),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: NxTypography.bodySm),
          Text(value, style: NxTypography.titleSm),
        ],
      ),
    );
  }
}
