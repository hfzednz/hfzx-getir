import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../routing/route_names.dart';
import '../../../../shared/business_rules/handoff_rules.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/scan_fab.dart';
import '../../domain/entities/handoff_task.dart';
import '../providers/dispatch_providers.dart';

class HandoffDetailScreen extends ConsumerWidget {
  const HandoffDetailScreen({super.key, required this.handoffId});
  final String handoffId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(handoffProvider(handoffId));
    final actions = ref.read(dispatchActionsProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Handoff'),
      floatingActionButton: async.maybeWhen(
        data: (h) => HandoffRules.canScanQr(h.status)
            ? ScanFab(
                label: 'Scan QR',
                onPressed: () =>
                    context.push(RouteNames.dispatchScanPath(handoffId)),
              )
            : null,
        orElse: () => null,
      ),
      body: AsyncValueWidget<HandoffTask>(
        value: async,
        data: (h) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s3),
          children: [
            NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Order ${h.orderId}', style: NxTypography.titleMd),
                  Text('Status: ${h.status.wireName}', style: NxTypography.captionMd),
                  if (h.courierName != null)
                    Text('Courier: ${h.courierName}', style: NxTypography.bodySm),
                  Text('${h.bagCount} bags', style: NxTypography.bodySm),
                ],
              ),
            ),
            const SizedBox(height: NxSpacing.s3),
            if (HandoffRules.canMarkArrived(h.status))
              NxButton(
                label: 'Courier arrived',
                expand: true,
                onPressed: () => actions.markArrived(handoffId),
              ),
            if (HandoffRules.canConfirm(h.status)) ...[
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Confirm handoff',
                expand: true,
                onPressed: () => actions.confirm(handoffId),
              ),
            ],
            if (HandoffRules.canFail(h.status)) ...[
              const SizedBox(height: NxSpacing.s2),
              NxButton(
                label: 'Fail pickup',
                expand: true,
                variant: NxButtonVariant.destructive,
                onPressed: () => actions.fail(handoffId, reasonCode: 'courier_timeout'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
