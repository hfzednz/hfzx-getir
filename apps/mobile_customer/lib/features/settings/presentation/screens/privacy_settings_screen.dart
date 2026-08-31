import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class PrivacySettingsScreen extends StatelessWidget {
  const PrivacySettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return Scaffold(
      appBar: NxTopBar(title: l10n.privacy),
      body: ListView(
        children: [
          ListTile(
            leading: const Icon(Icons.shield_outlined),
            title: Text(l10n.privacyControls),
            subtitle: Text(l10n.privacySubtitle),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsPrivacyControls),
          ),
          const Divider(height: 1),
          ListTile(
            leading: const Icon(Icons.description_outlined),
            title: Text(l10n.privacyPolicy),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/legal/privacy'),
          ),
          ListTile(
            leading: const Icon(Icons.gavel_outlined),
            title: Text(l10n.termsOfService),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/legal/terms'),
          ),
          ListTile(
            leading: const Icon(Icons.cookie_outlined),
            title: Text(l10n.cookiePolicy),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push('/legal/cookies'),
          ),
        ],
      ),
    );
  }
}
