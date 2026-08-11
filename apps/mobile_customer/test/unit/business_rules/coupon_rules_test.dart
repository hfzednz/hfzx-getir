import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/shared/business_rules/coupon_rules.dart';

CouponDefinition _coupon({
  String code = 'SAVE10',
  int minOrderMinor = 0,
  DateTime? expiresAt,
  DateTime? startsAt,
  int? usageLimit,
  int usageCount = 0,
  bool stackable = false,
  bool firstOrderOnly = false,
  bool active = true,
  List<String> categories = const [],
}) =>
    CouponDefinition(
      code: code,
      discountType: CouponDiscountType.percent,
      discountValue: 10,
      minOrderMinor: minOrderMinor,
      expiresAt: expiresAt,
      startsAt: startsAt,
      usageLimit: usageLimit,
      usageCount: usageCount,
      stackable: stackable,
      firstOrderOnly: firstOrderOnly,
      active: active,
      applicableCategoryIds: categories,
    );

void main() {
  group('CouponRules.evaluateEligibility', () {
    test('passes for valid coupon', () {
      final violations = CouponRules.evaluateEligibility(
        coupon: _coupon(),
        cartSubtotalMinor: 5000,
        cartCurrency: 'TRY',
        hasActivePromotion: false,
        isFirstOrder: true,
        cartCategoryIds: const {'grocery'},
      );
      expect(violations, isEmpty);
    });

    test('flags expired coupon', () {
      final violations = CouponRules.evaluateEligibility(
        coupon: _coupon(expiresAt: DateTime.utc(2020, 1, 1)),
        cartSubtotalMinor: 5000,
        cartCurrency: 'TRY',
        hasActivePromotion: false,
        isFirstOrder: true,
        cartCategoryIds: const {},
        referenceTime: DateTime.utc(2026, 1, 1),
      );
      expect(
        violations.any((v) => v.code == CouponViolationCode.expired),
        isTrue,
      );
    });

    test('flags min order not met', () {
      final violations = CouponRules.evaluateEligibility(
        coupon: _coupon(minOrderMinor: 10000),
        cartSubtotalMinor: 5000,
        cartCurrency: 'TRY',
        hasActivePromotion: false,
        isFirstOrder: true,
        cartCategoryIds: const {},
      );
      expect(
        violations.any((v) => v.code == CouponViolationCode.minOrderNotMet),
        isTrue,
      );
    });

    test('flags non-stackable coupon with active promotion', () {
      final violations = CouponRules.evaluateEligibility(
        coupon: _coupon(stackable: false),
        cartSubtotalMinor: 5000,
        cartCurrency: 'TRY',
        hasActivePromotion: true,
        isFirstOrder: true,
        cartCategoryIds: const {},
      );
      expect(
        violations.any((v) => v.code == CouponViolationCode.notStackable),
        isTrue,
      );
    });

    test('flags first-order-only for returning customer', () {
      final violations = CouponRules.evaluateEligibility(
        coupon: _coupon(firstOrderOnly: true),
        cartSubtotalMinor: 5000,
        cartCurrency: 'TRY',
        hasActivePromotion: false,
        isFirstOrder: false,
        cartCategoryIds: const {},
      );
      expect(
        violations.any((v) => v.code == CouponViolationCode.firstOrderOnly),
        isTrue,
      );
    });
  });

  group('CouponRules.computeDiscountMinor', () {
    test('computes percent discount capped by max', () {
      final discount = CouponRules.computeDiscountMinor(
        coupon: _coupon().copyWith(maxDiscountMinor: 300),
        cartSubtotalMinor: 10000,
      );
      expect(discount, 300);
    });

    test('computes fixed discount', () {
      final discount = CouponRules.computeDiscountMinor(
        coupon: const CouponDefinition(
          code: 'FLAT500',
          discountType: CouponDiscountType.fixedMinor,
          discountValue: 500,
        ),
        cartSubtotalMinor: 10000,
      );
      expect(discount, 500);
    });
  });
}

extension on CouponDefinition {
  CouponDefinition copyWith({int? maxDiscountMinor}) => CouponDefinition(
        code: code,
        discountType: discountType,
        discountValue: discountValue,
        minOrderMinor: minOrderMinor,
        maxDiscountMinor: maxDiscountMinor ?? this.maxDiscountMinor,
        expiresAt: expiresAt,
        startsAt: startsAt,
        usageLimit: usageLimit,
        usageCount: usageCount,
        stackable: stackable,
        firstOrderOnly: firstOrderOnly,
        currency: currency,
        active: active,
        applicableCategoryIds: applicableCategoryIds,
      );
}
