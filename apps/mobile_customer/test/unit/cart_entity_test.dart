import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/features/cart/data/local/app_database.dart';
import 'package:nexora_customer/shared/utils/money.dart';

void main() {
  group('Money', () {
    test('formats TRY minor units', () {
      const money = Money(minorUnits: 1250, currency: 'TRY');
      expect(money.majorUnits, 12.5);
      expect(money.format(), '₺12.50');
    });

    test('fromJson parses snake_case fields', () {
      final money = Money.fromJson({'minor_units': 999, 'currency': 'TRY'});
      expect(money.minorUnits, 999);
      expect(money.currency, 'TRY');
    });
  });

  group('CartItem entity', () {
    test('quantity defaults from companion', () {
      final item = CartItem(
        productId: 'p1',
        variantId: null,
        title: 'Milk',
        imageUrl: null,
        quantity: 2,
        unitPriceMinor: 500,
        currency: 'TRY',
        notes: null,
        updatedAt: DateTime.utc(2026, 1, 1),
        pendingSync: false,
      );
      expect(item.quantity, 2);
      expect(item.unitPriceMinor * item.quantity, 1000);
    });
  });
}
