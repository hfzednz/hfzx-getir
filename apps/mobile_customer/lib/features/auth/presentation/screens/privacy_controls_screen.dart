import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/privacy_controls.dart';
import '../providers/account_lifecycle_providers.dart';

class PrivacyControlsScreen extends ConsumerStatefulWidget {
  const PrivacyControlsScreen({super.key});

  @override
  ConsumerState<PrivacyControlsScreen> createState() =>
      _PrivacyControlsScreenState();
}

class _PrivacyControlsScreenState extends ConsumerState<PrivacyControlsScreen> {
  PrivacyControls? _draft;

  Future<void> _save() async {
    final controls = _draft;
    if (controls == null) return;
    final ok = await ref
        .read(accountLifecycleControllerProvider.notifier)
        .updatePrivacy(controls);
    if (!mounted) return;
    if (ok) {
      ref.invalidate(privacyControlsProvider);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).privacySettingsSaved)),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final controlsAsync = ref.watch(privacyControlsProvider);
    final lifecycle = ref.watch(accountLifecycleControllerProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.privacy),
      body: controlsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(l10n.emptyMessage),
              const SizedBox(height: NxSpacing.s3),
              NxButton(
                label: l10n.retry,
                onPressed: () => ref.invalidate(privacyControlsProvider),
              ),
            ],
          ),
        ),
        data: (controls) {
          final draft = _draft ?? controls;
          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              SwitchListTile(
                title: Text(l10n.marketingEmail),
                value: draft.marketingEmail,
                onChanged: (v) =>
                    setState(() => _draft = draft.copyWith(marketingEmail: v)),
              ),
              SwitchListTile(
                title: Text(l10n.pushNotifications),
                value: draft.marketingPush,
                onChanged: (v) =>
                    setState(() => _draft = draft.copyWith(marketingPush: v)),
              ),
              SwitchListTile(
                title: Text(l10n.smsMarketing),
                value: draft.marketingSms,
                onChanged: (v) =>
                    setState(() => _draft = draft.copyWith(marketingSms: v)),
              ),
              SwitchListTile(
                title: Text(l10n.personalization),
                value: draft.personalization,
                onChanged: (v) =>
                    setState(() => _draft = draft.copyWith(personalization: v)),
              ),
              SwitchListTile(
                title: Text(l10n.optOutAnalytics),
                value: draft.analyticsOptOut,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(analyticsOptOut: v),
                ),
              ),
              SwitchListTile(
                title: Text(l10n.shareWithPartners),
                value: draft.shareWithPartners,
                onChanged: (v) => setState(
                  () => _draft = draft.copyWith(shareWithPartners: v),
                ),
              ),
              const SizedBox(height: NxSpacing.s4),
              NxButton(
                label: l10n.save,
                expand: true,
                loading: lifecycle.isLoading,
                onPressed: _save,
              ),
            ],
          );
        },
      ),
    );
  }
}
