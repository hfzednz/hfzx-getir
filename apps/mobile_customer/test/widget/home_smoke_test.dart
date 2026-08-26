import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/l10n/app_localizations.dart';
import 'package:nexora_customer/shared/widgets/error_view.dart';
import 'package:nexora_design/nexora_design.dart';

/// Full [NexoraApp] starts splash → postBootstrap (Hive/Firebase/FCM/sync).
/// That never idles on the VM and trips the default 10-minute test timeout
/// (verified on 2fd4cb5 / 1eb3b25 rc-flutter-static).
void main() {
  testWidgets('home-adjacent chrome mounts', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        locale: const Locale('en'),
        localizationsDelegates: const [
          AppLocalizations.delegate,
          GlobalMaterialLocalizations.delegate,
          GlobalWidgetsLocalizations.delegate,
        ],
        supportedLocales: const [Locale('en'), Locale('tr')],
        theme: NxTheme.light(),
        home: const Scaffold(
          body: ErrorView(
            title: 'Popular',
            message: 'Test Product',
          ),
        ),
      ),
    );
    expect(find.text('Popular'), findsOneWidget);
    expect(find.text('Test Product'), findsOneWidget);
  }, timeout: const Timeout(Duration(seconds: 15)));
}
