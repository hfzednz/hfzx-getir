import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/profile_entity.dart';
import '../providers/profile_providers.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  final _firstName = TextEditingController();
  final _lastName = TextEditingController();
  final _displayName = TextEditingController();
  final _email = TextEditingController();
  final _phone = TextEditingController();
  bool _emailMarketing = true;
  bool _pushOrderUpdates = true;
  bool _pushPromotions = false;
  bool _smsAlerts = false;
  UserProfile? _loaded;

  @override
  void dispose() {
    _firstName.dispose();
    _lastName.dispose();
    _displayName.dispose();
    _email.dispose();
    _phone.dispose();
    super.dispose();
  }

  void _bind(UserProfile p) {
    if (_loaded?.id == p.id) return;
    _loaded = p;
    _firstName.text = p.firstName;
    _lastName.text = p.lastName;
    _displayName.text = p.displayName;
    _email.text = p.email;
    _phone.text = p.phone;
    _emailMarketing = p.communicationPreferences.emailMarketing;
    _pushOrderUpdates = p.communicationPreferences.pushOrderUpdates;
    _pushPromotions = p.communicationPreferences.pushPromotions;
    _smsAlerts = p.communicationPreferences.smsAlerts;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final profileAsync = ref.watch(userProfileProvider);
    final saveState = ref.watch(profileUpdateProvider);

    return Scaffold(
      appBar: NxTopBar(
        title: l10n.profileTitle,
        actions: [
          TextButton(
            onPressed: saveState.isLoading ? null : _save,
            child: const Text('Save'),
          ),
        ],
      ),
      body: AsyncValueWidget(
        value: profileAsync,
        data: (profile) {
          _bind(profile);
          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              TextField(controller: _firstName, decoration: const InputDecoration(labelText: 'First name')),
              const SizedBox(height: NxSpacing.s3),
              TextField(controller: _lastName, decoration: const InputDecoration(labelText: 'Last name')),
              const SizedBox(height: NxSpacing.s3),
              TextField(controller: _displayName, decoration: const InputDecoration(labelText: 'Display name')),
              const SizedBox(height: NxSpacing.s3),
              TextField(controller: _email, decoration: const InputDecoration(labelText: 'Email')),
              const SizedBox(height: NxSpacing.s3),
              TextField(controller: _phone, decoration: const InputDecoration(labelText: 'Phone')),
              const SizedBox(height: NxSpacing.s4),
              Text('Communication preferences', style: NxTypography.headlineSm),
              SwitchListTile(
                title: const Text('Email marketing'),
                value: _emailMarketing,
                onChanged: (v) => setState(() => _emailMarketing = v),
              ),
              SwitchListTile(
                title: const Text('Order updates (push)'),
                value: _pushOrderUpdates,
                onChanged: (v) => setState(() => _pushOrderUpdates = v),
              ),
              SwitchListTile(
                title: const Text('Promotions (push)'),
                value: _pushPromotions,
                onChanged: (v) => setState(() => _pushPromotions = v),
              ),
              SwitchListTile(
                title: const Text('SMS alerts'),
                value: _smsAlerts,
                onChanged: (v) => setState(() => _smsAlerts = v),
              ),
              const SizedBox(height: NxSpacing.s3),
              ListTile(
                title: const Text('Privacy controls'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () => context.push(RouteNames.settingsPrivacyControls),
              ),
              if (saveState.hasError)
                Padding(
                  padding: const EdgeInsets.only(top: NxSpacing.s3),
                  child: ErrorView(message: saveState.error.toString(), onRetry: _save),
                ),
            ],
          );
        },
        error: (e, _) => ErrorView(message: e.toString(), onRetry: () => ref.invalidate(userProfileProvider)),
      ),
    );
  }

  Future<void> _save() async {
    final base = _loaded;
    if (base == null) return;
    final first = _firstName.text.trim();
    final last = _lastName.text.trim();
    if (first.isEmpty || last.isEmpty) {
      final tr = Localizations.localeOf(context).languageCode == 'tr';
      NxToast.show(
        context,
        message: tr ? 'Ad ve soyad gerekli.' : 'First and last name are required.',
      );
      return;
    }
    await ref.read(profileUpdateProvider.notifier).save(
          base.copyWith(
            firstName: first,
            lastName: last,
            displayName: _displayName.text.trim(),
            email: _email.text.trim(),
            phone: _phone.text.trim(),
            communicationPreferences: CommunicationPreferences(
              emailMarketing: _emailMarketing,
              pushOrderUpdates: _pushOrderUpdates,
              pushPromotions: _pushPromotions,
              smsAlerts: _smsAlerts,
            ),
          ),
        );
    if (!mounted) return;
    final saveState = ref.read(profileUpdateProvider);
    if (!saveState.hasError) {
      final tr = Localizations.localeOf(context).languageCode == 'tr';
      NxToast.show(
        context,
        message: tr ? 'Profil kaydedildi.' : 'Profile saved.',
      );
      ref.invalidate(userProfileProvider);
    }
  }
}
