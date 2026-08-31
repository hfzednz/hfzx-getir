import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/errors/error_copy.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../notifications/domain/entities/notifications_entity.dart';
import '../../../notifications/presentation/providers/notifications_providers.dart';

class NotificationPreferencesScreen extends ConsumerStatefulWidget {
  const NotificationPreferencesScreen({super.key});

  @override
  ConsumerState<NotificationPreferencesScreen> createState() =>
      _NotificationPreferencesScreenState();
}

class _NotificationPreferencesScreenState
    extends ConsumerState<NotificationPreferencesScreen> {
  NotificationPreferences? _draft;
  bool _saving = false;

  Future<void> _save() async {
    final draft = _draft;
    if (draft == null) return;

    setState(() => _saving = true);
    final result =
        await ref.read(updateNotificationPreferencesUseCaseProvider).call(draft);
    if (!mounted) return;
    setState(() => _saving = false);

    result.fold(
      onSuccess: (prefs) {
        setState(() => _draft = prefs);
        ref.invalidate(notificationPreferencesProvider);
        NxToast.show(context, message: AppLocalizations.of(context).preferencesSaved);
      },
      onFailure: (e) => NxToast.show(
        context,
        message: localizedCustomerError(context, e),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final prefsAsync = ref.watch(notificationPreferencesProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.notificationPreferences),
      body: AsyncValueWidget(
        value: prefsAsync,
        data: (prefs) {
          _draft ??= prefs;
          final draft = _draft!;

          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              SwitchListTile(
                title: Text(l10n.pushNotifications),
                value: draft.pushEnabled,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(pushEnabled: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.emailNotifications),
                value: draft.emailEnabled,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(emailEnabled: v),
                ),
              ),
              const Divider(),
              SwitchListTile(
                title: Text(l10n.orderUpdates),
                subtitle: Text(l10n.transactional),
                value: draft.transactional,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(transactional: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.deliveryAlerts),
                value: draft.delivery,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(delivery: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.promotionsPref),
                value: draft.promo,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(promo: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.priceDrops),
                value: draft.priceDrop,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(priceDrop: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.backInStock),
                value: draft.restock,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(restock: v),
                ),
              ),
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: l10n.savePreferences,
                expand: true,
                loading: _saving,
                disabled: _saving,
                onPressed: _save,
              ),
            ],
          );
        },
        error: (e, _) => ErrorView(
          message: localizedCustomerError(context, e),
          onRetry: () => ref.invalidate(notificationPreferencesProvider),
        ),
      ),
    );
  }
}
