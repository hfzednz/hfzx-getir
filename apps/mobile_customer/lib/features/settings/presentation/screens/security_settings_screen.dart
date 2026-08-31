import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../auth/presentation/providers/auth_repository_provider.dart';

class SecuritySettingsScreen extends ConsumerStatefulWidget {
  const SecuritySettingsScreen({super.key});

  @override
  ConsumerState<SecuritySettingsScreen> createState() =>
      _SecuritySettingsScreenState();
}

class _SecuritySettingsScreenState
    extends ConsumerState<SecuritySettingsScreen> {
  bool? _biometricOverride;

  bool get _biometricEnabled =>
      _biometricOverride ??
      ref.read(preferencesStoreProvider).get<bool>('biometric_enabled') ??
      false;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return Scaffold(
      appBar: NxTopBar(title: l10n.security),
      body: ListView(
        children: [
          SwitchListTile(
            title: Text(l10n.biometricUnlock),
            subtitle: Text(l10n.biometricHint),
            value: _biometricEnabled,
            onChanged: (enabled) async {
              final repo = ref.read(authRepositoryProvider);
              if (enabled) {
                await repo.enableBiometric();
              } else {
                await repo.clearBiometric();
              }
              setState(() => _biometricOverride = enabled);
            },
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.devices_outlined),
            title: Text(l10n.trustedDevices),
            subtitle: Text(l10n.trustedDevicesHint),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsDevices),
          ),
          ListTile(
            leading: Icon(Icons.delete_forever_outlined, color: context.nxColors.danger),
            title: Text(
              l10n.deleteAccount,
              style: TextStyle(color: context.nxColors.danger),
            ),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsDeleteAccount),
          ),
        ],
      ),
    );
  }
}
