import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final themeMode = ref.watch(themeModeProvider);
    final locale = ref.watch(localeCodeProvider);

    final themeLabel = switch (themeMode) {
      ThemeMode.system => l10n.themeSystem,
      ThemeMode.light => l10n.themeLight,
      ThemeMode.dark => l10n.themeDark,
    };

    final languageLabel = locale == 'tr' ? 'Türkçe' : 'English';

    return Scaffold(
      appBar: NxTopBar(title: l10n.settingsTitle),
      body: ListView(
        children: [
          ListTile(
            leading: const Icon(Icons.language_outlined),
            title: Text(l10n.language),
            subtitle: Text(languageLabel),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsLanguage),
          ),
          ListTile(
            leading: const Icon(Icons.palette_outlined),
            title: Text(l10n.theme),
            subtitle: Text(themeLabel),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsTheme),
          ),
          ListTile(
            leading: const Icon(Icons.accessibility_new_outlined),
            title: const Text('Accessibility'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsA11y),
          ),
          ListTile(
            leading: const Icon(Icons.notifications_outlined),
            title: const Text('Notifications'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsNotifications),
          ),
          ListTile(
            leading: const Icon(Icons.privacy_tip_outlined),
            title: Text(l10n.privacy),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsPrivacy),
          ),
          ListTile(
            leading: const Icon(Icons.security_outlined),
            title: Text(l10n.security),
            trailing: const Icon(Icons.chevron_right),
            onTap: () => context.push(RouteNames.settingsSecurity),
          ),
        ],
      ),
    );
  }
}
