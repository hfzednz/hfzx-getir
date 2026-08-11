import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';

class AccessibilitySettingsScreen extends ConsumerStatefulWidget {
  const AccessibilitySettingsScreen({super.key});

  @override
  ConsumerState<AccessibilitySettingsScreen> createState() =>
      _AccessibilitySettingsScreenState();
}

class _AccessibilitySettingsScreenState
    extends ConsumerState<AccessibilitySettingsScreen> {
  static const _highContrastKey = 'a11y_high_contrast';
  static const _reducedMotionKey = 'a11y_reduced_motion';

  bool? _highContrast;
  bool? _reducedMotion;

  bool get _highContrastValue =>
      _highContrast ??
      ref.read(preferencesStoreProvider).get<bool>(_highContrastKey) ??
      false;

  bool get _reducedMotionValue =>
      _reducedMotion ??
      ref.read(preferencesStoreProvider).get<bool>(_reducedMotionKey) ??
      false;

  Future<void> _setHighContrast(bool value) async {
    await ref.read(preferencesStoreProvider).set(_highContrastKey, value);
    setState(() => _highContrast = value);
  }

  Future<void> _setReducedMotion(bool value) async {
    await ref.read(preferencesStoreProvider).set(_reducedMotionKey, value);
    setState(() => _reducedMotion = value);
  }

  @override
  Widget build(BuildContext context) {
    final textScaler = MediaQuery.textScalerOf(context);

    return Scaffold(
      appBar: const NxTopBar(title: 'Accessibility'),
      body: ListView(
        children: [
          SwitchListTile(
            title: const Text('High contrast'),
            subtitle: const Text('Increase contrast for text and controls'),
            value: _highContrastValue,
            onChanged: _setHighContrast,
          ),
          SwitchListTile(
            title: const Text('Reduce motion'),
            subtitle: const Text('Prefer less animation where supported'),
            value: _reducedMotionValue,
            onChanged: _setReducedMotion,
          ),
          ListTile(
            title: const Text('Text size'),
            subtitle: Text(
              'Follows your device text scale '
              '(${textScaler.scale(1.0).toStringAsFixed(2)}×). '
              'Change it in system accessibility settings.',
            ),
          ),
        ],
      ),
    );
  }
}
