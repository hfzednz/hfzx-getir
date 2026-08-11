import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../routing/route_names.dart';

class MoreScreen extends ConsumerWidget {
  const MoreScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final links = <(IconData, String, String)>[
      (Icons.inventory_2_outlined, 'Inventory', RouteNames.inventory),
      (Icons.swap_horiz, 'Transfers', RouteNames.transfers),
      (Icons.event_busy_outlined, 'Expiry', RouteNames.expiry),
      (Icons.shopping_cart_outlined, 'Purchasing', RouteNames.purchasing),
      (Icons.assignment_return_outlined, 'Returns', RouteNames.returns),
      (Icons.verified_outlined, 'Quality', RouteNames.quality),
      (Icons.map_outlined, 'Warehouse map', RouteNames.map),
      (Icons.auto_awesome, 'AI assist', RouteNames.ai),
      (Icons.schedule, 'Shifts', RouteNames.shifts),
      (Icons.checklist_outlined, 'Tasks', RouteNames.tasks),
      (Icons.bar_chart_outlined, 'Reports', RouteNames.reports),
      (Icons.settings_outlined, 'Settings', RouteNames.settings),
      (Icons.support_agent_outlined, 'Support', RouteNames.support),
    ];

    return Scaffold(
      appBar: const NxTopBar(title: 'More'),
      body: ListView.separated(
        padding: const EdgeInsets.all(NxSpacing.s3),
        itemCount: links.length,
        separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
        itemBuilder: (context, i) {
          final (icon, label, path) = links[i];
          return NxCard(
            variant: NxCardVariant.interactive,
            onTap: () => context.push(path),
            child: Row(
              children: [
                Icon(icon, size: NxIconSize.md),
                const SizedBox(width: NxSpacing.s3),
                Expanded(child: Text(label, style: NxTypography.titleSm)),
                const Icon(Icons.chevron_right, size: NxIconSize.md),
              ],
            ),
          );
        },
      ),
    );
  }
}
