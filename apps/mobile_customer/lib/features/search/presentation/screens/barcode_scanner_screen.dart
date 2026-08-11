import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';

class BarcodeScannerScreen extends StatelessWidget {
  const BarcodeScannerScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Scaffold(
      appBar: NxTopBar(title: l10n.barcodeScan),
      body: MobileScanner(
        onDetect: (capture) {
          final code = capture.barcodes.firstOrNull?.rawValue;
          if (code != null) context.pop(code);
        },
      ),
    );
  }
}
