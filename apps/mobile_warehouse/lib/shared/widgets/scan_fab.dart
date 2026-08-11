import 'package:flutter/material.dart';
import 'package:nexora_design/nexora_design.dart';

/// Large scan CTA for dense warehouse ops screens.
class ScanFab extends StatelessWidget {
  const ScanFab({
    super.key,
    required this.onPressed,
    this.label = 'Scan',
  });

  final VoidCallback onPressed;
  final String label;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    return FloatingActionButton.extended(
      onPressed: onPressed,
      backgroundColor: colors.bgAccent,
      foregroundColor: colors.textOnAccent,
      icon: const Icon(Icons.qr_code_scanner, size: 28),
      label: Text(label, style: NxTypography.titleSm),
      extendedPadding: const EdgeInsets.symmetric(
        horizontal: NxSpacing.s6,
        vertical: NxSpacing.s4,
      ),
    );
  }
}
