import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/delivery_job.dart';
import '../providers/deliveries_providers.dart';

class DeliveriesScreen extends ConsumerWidget {
  const DeliveriesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(activeDeliveriesProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Deliveries'),
      body: AsyncValueWidget<List<DeliveryJob>>(
        value: async,
        data: (jobs) {
          if (jobs.isEmpty) {
            return const NxEmptyState(
              title: 'Nothing here yet',
              body: 'Check back soon.',
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(activeDeliveriesProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: jobs.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, i) {
                final job = jobs[i];
                final colors = context.nxColors;
                return NxCard(
                  onTap: () =>
                      context.push(RouteNames.deliveryDetailPath(job.id)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        job.storeName,
                        style: NxTypography.titleSm
                            .copyWith(color: colors.textPrimary),
                      ),
                      Text(
                        job.customerArea,
                        style: NxTypography.bodySm
                            .copyWith(color: colors.textSecondary),
                      ),
                      const SizedBox(height: NxSpacing.s1),
                      Text(
                        job.status.apiValue,
                        style: NxTypography.captionMd
                            .copyWith(color: colors.textBrand),
                      ),
                    ],
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }
}
