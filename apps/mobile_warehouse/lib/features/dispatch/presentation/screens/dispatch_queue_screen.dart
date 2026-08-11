import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/handoff_task.dart';
import '../providers/dispatch_providers.dart';

class DispatchQueueScreen extends ConsumerWidget {
  const DispatchQueueScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(handoffRealtimeInvalidationProvider);
    final async = ref.watch(dispatchQueueProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Dispatch'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(dispatchQueueProvider),
        child: AsyncValueWidget<List<HandoffTask>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [
                SizedBox(height: 120),
                NxEmptyState(title: 'No handoffs', body: 'Waiting for packed orders.'),
              ]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final h = items[i];
                return NxCard(
                  variant: NxCardVariant.interactive,
                  onTap: () => context.push(RouteNames.dispatchHandoffPath(h.id)),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Order ${h.orderId}', style: NxTypography.titleSm),
                      Text(
                        '${h.status.wireName}'
                        '${h.courierName != null ? ' · ${h.courierName}' : ''}'
                        ' · ${h.bagCount} bags',
                        style: NxTypography.captionMd,
                      ),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}
