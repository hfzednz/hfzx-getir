import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../status/presentation/widgets/status_control_widget.dart';
import '../../domain/entities/dashboard_entity.dart';
import '../providers/dashboard_providers.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dashAsync = ref.watch(dashboardProvider);
    final batteryLow = ref.watch(batteryLowWarningProvider);
    final connection = ref.watch(connectionQualityProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: NxTopBar(
        title: 'Home',
        actions: [
          IconButton(
            icon: const Icon(Icons.notifications_outlined),
            onPressed: () => context.push(RouteNames.notifications),
          ),
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => context.push(RouteNames.settings),
          ),
        ],
      ),
      body: AsyncValueWidget<CourierDashboard>(
        value: dashAsync,
        data: (dash) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(dashboardProvider),
          child: ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              const StatusControlWidget(),
              const SizedBox(height: NxSpacing.s3),
              if (batteryLow)
                Padding(
                  padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                  child: NxBanner(
                    title: 'Low battery',
                    message: 'Low battery — location updates slowed',
                    variant: NxBannerVariant.warning,
                  ),
                ),
              if (connection == 'offline' || connection == 'poor')
                Padding(
                  padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                  child: NxBanner(
                    title: 'Connection',
                    message: connection == 'offline'
                        ? 'You are offline — actions will sync when connected'
                        : 'Poor connection',
                    variant: NxBannerVariant.warning,
                  ),
                ),
              NxCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "Today's earnings",
                      style: NxTypography.captionMd
                          .copyWith(color: colors.textSecondary),
                    ),
                    Text(
                      Money(
                        minorUnits: dash.todayEarningsMinor,
                        currency: dash.currency,
                      ).format(),
                      style: NxTypography.dashKpi
                          .copyWith(color: colors.textPrimary),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: NxSpacing.s3),
              Row(
                children: [
                  Expanded(
                    child: _StatTile(
                      label: 'Completed',
                      value: '${dash.completedCount}',
                    ),
                  ),
                  const SizedBox(width: NxSpacing.s2),
                  Expanded(
                    child: _StatTile(
                      label: 'Pending',
                      value: '${dash.pendingCount}',
                    ),
                  ),
                ],
              ),
              const SizedBox(height: NxSpacing.s2),
              Row(
                children: [
                  Expanded(
                    child: _StatTile(
                      label: 'Acceptance',
                      value:
                          '${(dash.acceptanceRate * 100).toStringAsFixed(0)}%',
                    ),
                  ),
                  const SizedBox(width: NxSpacing.s2),
                  Expanded(
                    child: _StatTile(
                      label: 'Performance',
                      value: dash.performanceScore.toStringAsFixed(1),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: NxSpacing.s3),
              _ShiftCard(
                title: 'Current shift',
                shift: dash.currentShift,
              ),
              const SizedBox(height: NxSpacing.s2),
              _ShiftCard(
                title: 'Next shift',
                shift: dash.nextShift,
              ),
              if (dash.aiTip != null && dash.aiTip!.isNotEmpty) ...[
                const SizedBox(height: NxSpacing.s3),
                NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'AI tip',
                        style: NxTypography.titleSm
                            .copyWith(color: colors.textBrand),
                      ),
                      const SizedBox(height: NxSpacing.s1),
                      Text(
                        dash.aiTip!,
                        style: NxTypography.bodyMd
                            .copyWith(color: colors.textPrimary),
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: 'Offers',
                expand: true,
                onPressed: () => context.push(RouteNames.offers),
              ),
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Deliveries',
                variant: NxButtonVariant.secondary,
                expand: true,
                onPressed: () => context.push(RouteNames.deliveries),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatTile extends StatelessWidget {
  const _StatTile({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return NxCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
          ),
          Text(
            value,
            style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
          ),
        ],
      ),
    );
  }
}

class _ShiftCard extends StatelessWidget {
  const _ShiftCard({required this.title, required this.shift});
  final String title;
  final ShiftSummary? shift;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final label = shift?.label ??
        (shift?.startsAt != null
            ? '${shift!.startsAt} → ${shift!.endsAt ?? '—'}'
            : 'No shift');
    return NxCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
          ),
          Text(
            label,
            style: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
          ),
        ],
      ),
    );
  }
}
