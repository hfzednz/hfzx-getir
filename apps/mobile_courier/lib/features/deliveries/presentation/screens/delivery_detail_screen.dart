import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/delivery_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/delivery_job.dart';
import '../providers/deliveries_providers.dart';

class DeliveryDetailScreen extends ConsumerWidget {
  const DeliveryDetailScreen({super.key, required this.deliveryId});

  final String deliveryId;

  DeliveryLifecycleStatus? _nextStep(DeliveryLifecycleStatus status) {
    return switch (status) {
      DeliveryLifecycleStatus.assigned => DeliveryLifecycleStatus.enRouteStore,
      DeliveryLifecycleStatus.enRouteStore => DeliveryLifecycleStatus.atStore,
      DeliveryLifecycleStatus.atStore => null,
      DeliveryLifecycleStatus.pickedUp =>
        DeliveryLifecycleStatus.enRouteCustomer,
      DeliveryLifecycleStatus.enRouteCustomer =>
        DeliveryLifecycleStatus.arrived,
      DeliveryLifecycleStatus.arrived => null,
      _ => null,
    };
  }

  String _stepLabel(DeliveryLifecycleStatus status) => switch (status) {
        DeliveryLifecycleStatus.enRouteStore => 'En route to store',
        DeliveryLifecycleStatus.atStore => 'Arrived at store',
        DeliveryLifecycleStatus.enRouteCustomer => 'En route to customer',
        DeliveryLifecycleStatus.arrived => 'Arrived at customer',
        _ => status.apiValue,
      };

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(deliveryDetailProvider(deliveryId));
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Delivery'),
      body: AsyncValueWidget<DeliveryJob>(
        value: async,
        data: (job) {
          final next = _nextStep(job.status);
          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              NxCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      job.storeName,
                      style: NxTypography.titleMd
                          .copyWith(color: colors.textPrimary),
                    ),
                    Text(
                      job.customerArea,
                      style: NxTypography.bodyMd
                          .copyWith(color: colors.textSecondary),
                    ),
                    const SizedBox(height: NxSpacing.s2),
                    Text(
                      'Status: ${job.status.apiValue}',
                      style: NxTypography.captionMd
                          .copyWith(color: colors.textBrand),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: NxSpacing.s3),
              Text(
                'Workflow',
                style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
              ),
              const SizedBox(height: NxSpacing.s2),
              ...DeliveryLifecycleStatus.values
                  .where((s) =>
                      !DeliveryRules.isTerminal(s) ||
                      s == DeliveryLifecycleStatus.delivered)
                  .take(7)
                  .map(
                    (s) => Padding(
                      padding: const EdgeInsets.only(bottom: NxSpacing.s1),
                      child: Row(
                        children: [
                          Icon(
                            job.status.index >= s.index
                                ? Icons.check_circle
                                : Icons.radio_button_unchecked,
                            size: 18,
                            color: job.status.index >= s.index
                                ? colors.success
                                : colors.textTertiary,
                          ),
                          const SizedBox(width: NxSpacing.s2),
                          Text(
                            s.apiValue,
                            style: NxTypography.bodySm
                                .copyWith(color: colors.textPrimary),
                          ),
                        ],
                      ),
                    ),
                  ),
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: 'Navigate',
                expand: true,
                onPressed: () =>
                    context.push(RouteNames.deliveryNavigatePath(job.id)),
              ),
              if (DeliveryRules.canScanPickup(job.status)) ...[
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: 'Scan QR',
                  variant: NxButtonVariant.secondary,
                  expand: true,
                  onPressed: () =>
                      context.push(RouteNames.deliveryPickupPath(job.id)),
                ),
              ],
              if (next != null) ...[
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: _stepLabel(next),
                  expand: true,
                  onPressed: () async {
                    await ref.read(deliveryActionsProvider).advance(job, next);
                  },
                ),
              ],
              if (DeliveryRules.canSubmitPod(job.status)) ...[
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: 'Proof of delivery',
                  expand: true,
                  onPressed: () =>
                      context.push(RouteNames.deliveryPodPath(job.id)),
                ),
              ],
              if (DeliveryRules.canMarkFailed(job.status)) ...[
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: 'Mark as failed',
                  variant: NxButtonVariant.destructive,
                  expand: true,
                  onPressed: () =>
                      context.push(RouteNames.deliveryFailedPath(job.id)),
                ),
              ],
            ],
          );
        },
      ),
    );
  }
}
