import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:nexora_customer/main.dart' as app;

/// Device/emulator journey. GHA without an emulator marks Flutter UI checkout BLOCKED
/// and runs `test/live/bff_checkout_journey_test.dart` against the real BFF instead.
void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('app boots to splash then shows NEXORA', (tester) async {
    await app.main();
    await tester.pump(const Duration(milliseconds: 100));
    await tester.pumpAndSettle(const Duration(seconds: 8));
    expect(find.text('NEXORA'), findsWidgets);
  });
}
