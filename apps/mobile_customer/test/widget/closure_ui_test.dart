import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/l10n/app_localizations.dart';
import 'package:nexora_customer/shared/widgets/error_view.dart';
import 'package:nexora_design/nexora_design.dart';

Widget _app({required Locale locale, required Widget home}) {
  return MaterialApp(
    locale: locale,
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
      GlobalCupertinoLocalizations.delegate,
    ],
    supportedLocales: const [Locale('en'), Locale('tr')],
    theme: NxTheme.light(),
    home: home,
  );
}

void main() {
  testWidgets('quantity selector displays exactly 7', (tester) async {
    await tester.pumpWidget(
      _app(
        locale: const Locale('en'),
        home: const Scaffold(
          body: NxQtySelector(
            quantity: 7,
            onIncrement: _noop,
            onDecrement: _noop,
          ),
        ),
      ),
    );
    expect(find.text('7'), findsOneWidget);
  }, timeout: const Timeout(Duration(seconds: 15,)),);

  testWidgets('Turkish error view uses localized retry', (tester) async {
    await tester.pumpWidget(
      _app(
        locale: const Locale('tr'),
        home: const Scaffold(
          body: ErrorView(
            message: 'Bağlantı kurulamadı',
            onRetry: _noop,
          ),
        ),
      ),
    );
    expect(find.text('Bir sorun oluştu'), findsOneWidget);
    expect(find.text('Tekrar dene'), findsOneWidget);
    expect(find.text('Retry'), findsNothing);
  }, timeout: const Timeout(Duration(seconds: 15,)),);

  testWidgets('Turkish settings copy is generated', (tester) async {
    late AppLocalizations l10n;
    await tester.pumpWidget(
      _app(
        locale: const Locale('tr'),
        home: Builder(
          builder: (context) {
            l10n = AppLocalizations.of(context);
            return const SizedBox.shrink();
          },
        ),
      ),
    );
    expect(l10n.accessibility, 'Erişilebilirlik');
    expect(l10n.notificationsTitle, 'Bildirimler');
    expect(l10n.placeOrder, 'Siparişi ver');
    expect(l10n.apply, 'Uygula');
    expect(l10n.quantityLabel(7), 'Adet 7');
  }, timeout: const Timeout(Duration(seconds: 15,)),);
}

void _noop() {}
