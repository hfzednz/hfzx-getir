import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';

class ThemeSettingsScreen extends ConsumerWidget {
  const ThemeSettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final themeMode = ref.watch(themeModeProvider);

    final options = <(ThemeMode mode, String label)>[
      (ThemeMode.system, l10n.themeSystem),
      (ThemeMode.light, l10n.themeLight),
      (ThemeMode.dark, l10n.themeDark),
    ];

    return Scaffold(
      appBar: NxTopBar(title: l10n.theme),
      body: ListView(
        children: [
          for (final option in options)
            RadioListTile<ThemeMode>(
              title: Text(option.$2),
              value: option.$1,
              groupValue: themeMode,
              onChanged: (value) {
                if (value == null) return;
                ref.read(themeModeProvider.notifier).state = value;
                ref.read(preferencesStoreProvider).set('theme_mode', value.name);
              },
            ),
        ],
      ),
    );
  }
}
