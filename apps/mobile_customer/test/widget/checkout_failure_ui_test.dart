import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/l10n/app_localizations.dart';
import 'package:nexora_customer/shared/widgets/error_view.dart';
import 'package:nexora_design/nexora_design.dart';

Widget _wrap(Widget child) {
  return MaterialApp(
    locale: const Locale('en'),
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
    ],
    supportedLocales: const [Locale('en'), Locale('tr')],
    theme: NxTheme.light(),
    home: Scaffold(body: child),
  );
}

void main() {
  testWidgets('empty cart error is announced', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Your cart is empty',
          message: 'Add items before checkout.',
        ),
      ),
    );
    expect(find.text('Your cart is empty'), findsOneWidget);
    expect(find.text('Add items before checkout.'), findsOneWidget);
  });

  testWidgets('payment declined is announced', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Payment declined',
          message: 'The card was declined. No order was created.',
        ),
      ),
    );
    expect(find.textContaining('Payment declined'), findsOneWidget);
  });

  testWidgets('payment timeout is announced', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Payment timed out',
          message: 'Retry will reuse the same idempotency key.',
        ),
      ),
    );
    expect(find.textContaining('timed out'), findsOneWidget);
  });

  testWidgets('out of stock is announced', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Out of stock',
          message: 'This item is no longer available.',
        ),
      ),
    );
    expect(find.textContaining('Out of stock'), findsOneWidget);
  });

  testWidgets('invalid and expired promotion errors', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Promotion invalid',
          message: 'This code is expired or not applicable.',
        ),
      ),
    );
    expect(find.textContaining('Promotion invalid'), findsOneWidget);
  });

  testWidgets('session expiration error', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Session expired',
          message: 'Sign in again to continue checkout.',
        ),
      ),
    );
    expect(find.textContaining('Session expired'), findsOneWidget);
  });

  testWidgets('network interruption error', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Network unavailable',
          message: 'Reconnect and retry. Duplicate pay is blocked by idempotency.',
        ),
      ),
    );
    expect(find.textContaining('Network unavailable'), findsOneWidget);
  });

  testWidgets('server failure error', (tester) async {
    await tester.pumpWidget(
      _wrap(
        const ErrorView(
          title: 'Server error',
          message: 'Checkout was not completed.',
        ),
      ),
    );
    expect(find.textContaining('Server error'), findsOneWidget);
  });
}
