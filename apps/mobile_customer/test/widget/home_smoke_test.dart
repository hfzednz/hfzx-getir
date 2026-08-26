import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:nexora_customer/app/nexora_app.dart';
import 'package:nexora_customer/bootstrap/bootstrap.dart';
import 'package:nexora_customer/di/providers.dart';
import 'package:nexora_customer/features/home/domain/entities/home_entity.dart';
import 'package:nexora_customer/features/home/presentation/providers/home_providers.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('home screen smoke test', (tester) async {
    final bootstrapResult = await bootstrap();

    await tester.pumpWidget(
      ProviderScope(
        overrides: [
          ...bootstrapResult.overrides,
          environmentProvider.overrideWithValue(const AppEnvironment.dev()),
          connectivityOnlineProvider.overrideWith((ref) async* {
            yield true;
          }),
          homeFeedProvider.overrideWith(
            (ref) async => HomeFeed(
              widgets: [
                HomeWidgetConfig(
                  id: 'popular',
                  type: HomeWidgetType.trending,
                  title: 'Popular',
                  items: const [
                    HomeProduct(
                      id: '1',
                      title: 'Test Product',
                      priceMinor: 1999,
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
        child: const NexoraApp(),
      ),
    );

    // Connectivity/realtime streams never go idle; do not pumpAndSettle.
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 400));

    expect(find.byType(NexoraApp), findsOneWidget);
  });
}
