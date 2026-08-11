import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:local_auth/local_auth.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(localeCodeProvider);
    final themeMode = ref.watch(themeModeProvider);
    final biometric = ref.watch(biometricEnabledProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Settings'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text(
            'Language',
            style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
          ),
          const SizedBox(height: NxSpacing.s2),
          Row(
            children: [
              Expanded(
                child: NxButton(
                  label: 'English',
                  size: NxButtonSize.sm,
                  variant: locale == 'en'
                      ? NxButtonVariant.primary
                      : NxButtonVariant.secondary,
                  onPressed: () =>
                      ref.read(localeCodeProvider.notifier).state = 'en',
                ),
              ),
              const SizedBox(width: NxSpacing.s2),
              Expanded(
                child: NxButton(
                  label: 'Türkçe',
                  size: NxButtonSize.sm,
                  variant: locale == 'tr'
                      ? NxButtonVariant.primary
                      : NxButtonVariant.secondary,
                  onPressed: () =>
                      ref.read(localeCodeProvider.notifier).state = 'tr',
                ),
              ),
            ],
          ),
          const SizedBox(height: NxSpacing.s4),
          Text(
            'Theme',
            style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
          ),
          const SizedBox(height: NxSpacing.s2),
          ...ThemeMode.values.map(
            (mode) => RadioListTile<ThemeMode>(
              value: mode,
              groupValue: themeMode,
              title: Text(mode.name),
              onChanged: (v) {
                if (v != null) {
                  ref.read(themeModeProvider.notifier).state = v;
                }
              },
            ),
          ),
          const SizedBox(height: NxSpacing.s2),
          SwitchListTile(
            title: const Text('Biometric unlock'),
            value: biometric,
            onChanged: (enabled) async {
              if (enabled) {
                final auth = LocalAuthentication();
                final ok = await auth.authenticate(
                  localizedReason: 'Enable biometric unlock',
                );
                if (!ok) return;
              }
              ref.read(biometricEnabledProvider.notifier).state = enabled;
              await ref.read(preferencesStoreProvider).set(
                    'biometric_enabled',
                    enabled,
                  );
            },
          ),
        ],
      ),
    );
  }
}
