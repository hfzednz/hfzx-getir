import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/notification_entity.dart';
import '../providers/notifications_providers.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(notificationsProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Notifications'),
      body: AsyncValueWidget<List<CourierNotification>>(
        value: async,
        data: (items) {
          if (items.isEmpty) {
            return const NxEmptyState(
              title: 'Nothing here yet',
              body: 'Check back soon.',
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(notificationsProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final n = items[i];
                return NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        n.title,
                        style: NxTypography.titleSm.copyWith(
                          color: n.read
                              ? colors.textSecondary
                              : colors.textPrimary,
                        ),
                      ),
                      Text(
                        n.body,
                        style: NxTypography.bodySm
                            .copyWith(color: colors.textSecondary),
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
