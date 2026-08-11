import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/notifications_entity.dart';
import '../providers/notifications_providers.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(notificationsListProvider);

    return Scaffold(
      appBar: NxTopBar(
        title: l10n.notificationsTitle,
        actions: [
          NxIconButton(
            icon: const Icon(Icons.done_all),
            semanticLabel: 'Mark all read',
            onPressed: () async {
              await ref.read(markAllNotificationsReadUseCaseProvider).call();
              ref.invalidate(notificationsListProvider);
            },
          ),
        ],
      ),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(notificationsListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(notificationsListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                return NxCard(
                  variant: item.read ? NxCardVariant.staticCard : NxCardVariant.elevated,
                  child: ListTile(
                    leading: Icon(
                      _iconForType(item.type),
                      color: item.read
                          ? context.nxColors.iconSecondary
                          : context.nxColors.iconBrand,
                    ),
                    title: Text(
                      item.title.isNotEmpty ? item.title : item.id,
                      style: item.read
                          ? null
                          : NxTypography.titleSm.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                    ),
                    subtitle: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (item.body.isNotEmpty) Text(item.body),
                        if (item.createdAt != null)
                          Text(
                            _formatWhen(item.createdAt!),
                            style: NxTypography.captionSm.copyWith(
                              color: context.nxColors.textSecondary,
                            ),
                          ),
                      ],
                    ),
                    trailing: item.read
                        ? const Icon(Icons.chevron_right)
                        : Container(
                            width: 8,
                            height: 8,
                            decoration: BoxDecoration(
                              color: context.nxColors.bgBrand,
                              shape: BoxShape.circle,
                            ),
                          ),
                    onTap: () => _openNotification(context, ref, item),
                  ),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate(notificationsListProvider),
        ),
      ),
    );
  }

  IconData _iconForType(NotificationType type) => switch (type) {
        NotificationType.transactional => Icons.receipt_long_outlined,
        NotificationType.promo => Icons.local_offer_outlined,
        NotificationType.delivery => Icons.delivery_dining_outlined,
        NotificationType.priceDrop => Icons.trending_down,
        NotificationType.restock => Icons.inventory_2_outlined,
      };

  String _formatWhen(DateTime when) {
    final local = when.toLocal();
    return '${local.day}/${local.month}/${local.year} '
        '${local.hour.toString().padLeft(2, '0')}:'
        '${local.minute.toString().padLeft(2, '0')}';
  }

  Future<void> _openNotification(
    BuildContext context,
    WidgetRef ref,
    AppNotification item,
  ) async {
    if (!item.read) {
      await ref.read(markNotificationReadUseCaseProvider).call(item.id);
      ref.invalidate(notificationsListProvider);
    }
    if (item.deepLink != null && item.deepLink!.isNotEmpty && context.mounted) {
      context.push(item.deepLink!);
    }
  }
}
