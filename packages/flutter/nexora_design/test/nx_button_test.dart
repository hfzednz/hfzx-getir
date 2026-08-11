import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_design/nexora_design.dart';

void main() {
  group('NxButton', () {
    testWidgets('renders label and fires onPressed', (tester) async {
      var pressed = false;

      await tester.pumpWidget(
        MaterialApp(
          theme: NxTheme.light(),
          home: Scaffold(
            body: NxButton(
              label: 'Continue',
              onPressed: () => pressed = true,
            ),
          ),
        ),
      );

      expect(find.text('Continue'), findsOneWidget);
      expect(find.byType(NxButton), findsOneWidget);

      await tester.tap(find.byType(NxButton));
      await tester.pump();

      expect(pressed, isTrue);
    });

    testWidgets('shows spinner when loading and disables tap', (tester) async {
      var pressed = false;

      await tester.pumpWidget(
        MaterialApp(
          theme: NxTheme.light(),
          home: Scaffold(
            body: NxButton(
              label: 'Submit',
              loading: true,
              onPressed: () => pressed = true,
            ),
          ),
        ),
      );

      expect(find.byType(NxSpinner), findsOneWidget);
      expect(find.text('Submit'), findsNothing);

      await tester.tap(find.byType(NxButton));
      await tester.pump();

      expect(pressed, isFalse);
    });

    testWidgets('uses radius.md not pill shape', (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: NxTheme.light(),
          home: const Scaffold(
            body: NxButton(label: 'Primary'),
          ),
        ),
      );

      final material = tester.widget<Material>(find.descendant(
        of: find.byType(NxButton),
        matching: find.byType(Material),
      ));

      expect(
        material.borderRadius,
        BorderRadius.circular(NxRadius.md),
      );
    });
  });
}
