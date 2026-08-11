import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/features/cart/domain/entities/cart_entity.dart';
import 'package:nexora_customer/features/product/domain/entities/product_entity.dart';
import 'package:nexora_customer/shared/business_rules/cart_rules.dart';

Cart _cart({
  String id = 'c1',
  List<CartLine>? items,
  int subtotalMinor = 0,
  int? minOrderMinor,
  CartCoupon? coupon,
}) =>
    Cart(
      id: id,
      items: items ?? const [],
      subtotalMinor: subtotalMinor,
      totalMinor: subtotalMinor,
      minOrderMinor: minOrderMinor,
      coupon: coupon,
    );

CartLine _line({
  String productId = 'p1',
  String title = 'Milk',
  int quantity = 1,
  int unitPriceMinor = 500,
  int? maxQty,
  int? availableQty,
  ProductStockStatus stock = ProductStockStatus.inStock,
}) =>
    CartLine(
      productId: productId,
      title: title,
      quantity: quantity,
      unitPriceMinor: unitPriceMinor,
      maxQty: maxQty,
      availableQty: availableQty,
      stockStatus: stock,
    );

void main() {
  group('CartRules.evaluate', () {
    test('flags min order violation', () {
      final violations = CartRules.evaluate(
        _cart(
          items: [_line()],
          subtotalMinor: 1000,
          minOrderMinor: 2000,
        ),
      );
      expect(violations.any((v) => v.code == CartViolationCode.minOrder), isTrue);
    });

    test('flags out of stock line', () {
      final violations = CartRules.evaluate(
        _cart(items: [_line(stock: ProductStockStatus.outOfStock)]),
      );
      expect(violations.any((v) => v.code == CartViolationCode.outOfStock), isTrue);
    });

    test('flags max quantity from line metadata', () {
      final violations = CartRules.evaluate(
        _cart(items: [_line(quantity: 5, maxQty: 3)]),
      );
      expect(violations.any((v) => v.code == CartViolationCode.maxQty), isTrue);
    });

    test('flags cart weight limit from product meta', () {
      final violations = CartRules.evaluate(
        _cart(
          items: [_line(productId: 'heavy', quantity: 2)],
          subtotalMinor: 1000,
        ),
        context: const CartRulesContext(
          maxCartWeightGrams: 1000,
          productMeta: {
            'heavy': CartProductMeta(weightGrams: 600),
          },
        ),
      );
      expect(violations.any((v) => v.code == CartViolationCode.maxWeight), isTrue);
    });

    test('flags age restricted items when user not verified', () {
      final violations = CartRules.evaluate(
        _cart(items: [_line(productId: 'wine')]),
        context: const CartRulesContext(
          userAgeVerified: false,
          productMeta: { 'wine': CartProductMeta(ageRestricted: true) },
        ),
      );
      expect(violations.any((v) => v.code == CartViolationCode.ageRestricted), isTrue);
    });

    test('flags invalid coupon', () {
      final violations = CartRules.evaluate(
        _cart(
          coupon: const CartCoupon(code: 'BAD', discountMinor: 100, valid: false),
        ),
      );
      expect(violations.any((v) => v.code == CartViolationCode.couponInvalid), isTrue);
    });
  });

  group('CartRules.mergeAnonymousAndCloud', () {
    test('merges quantities for same product key', () {
      final local = _cart(
        id: 'local',
        items: [_line(productId: 'p1', quantity: 2)],
        subtotalMinor: 1000,
      );
      final cloud = _cart(
        id: 'cloud',
        items: [_line(productId: 'p1', quantity: 3, unitPriceMinor: 600)],
        subtotalMinor: 1500,
      );

      final result = CartRules.mergeAnonymousAndCloud(local: local, cloud: cloud);
      expect(result.isSuccess, isTrue);
      final merged = result.valueOrNull!;
      expect(merged.id, 'cloud');
      expect(merged.items.single.quantity, 5);
      expect(merged.items.single.unitPriceMinor, 600);
    });

    test('returns cloud cart when local is empty', () {
      final cloud = _cart(items: [_line()], subtotalMinor: 500);
      final result = CartRules.mergeAnonymousAndCloud(
        local: _cart(),
        cloud: cloud,
      );
      expect(result.isSuccess, isTrue);
      expect(result.valueOrNull!.items.length, 1);
    });
  });

  group('CartRules.validateForCheckout', () {
    test('returns failure when blocking violations exist', () {
      final result = CartRules.validateForCheckout(
        _cart(
          items: [_line(stock: ProductStockStatus.outOfStock)],
          subtotalMinor: 500,
        ),
      );
      expect(result.isFailure, isTrue);
    });

    test('returns success for valid cart', () {
      final result = CartRules.validateForCheckout(
        _cart(items: [_line()], subtotalMinor: 5000, minOrderMinor: 1000),
      );
      expect(result.isSuccess, isTrue);
    });
  });
}
