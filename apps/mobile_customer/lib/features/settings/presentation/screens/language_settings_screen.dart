import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';

class LanguageSettingsScreen extends ConsumerWidget {
  const LanguageSettingsScreen({super.key});

  static const _options = [
    (code: 'en', label: 'English'),
    (code: 'tr', label: 'Türkçe'),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final locale = ref.watch(localeCodeProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.language),
      body: RadioGroup<String>(
        groupValue: locale,
        onChanged: (value) {
          if (value == null) return;
          ref.read(localeCodeProvider.notifier).state = value;
          ref.read(preferencesStoreProvider).set('locale_code', value);
        },
        child: ListView(
          children: [
            for (final option in _options)
              RadioListTile<String>(
                title: Text(option.label),
                subtitle: Text(option.code.toUpperCase()),
                value: option.code,
              ),
          ],
        ),
      ),
    );
  }
}
