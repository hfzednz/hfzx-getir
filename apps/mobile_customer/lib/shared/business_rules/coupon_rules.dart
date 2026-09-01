import 'package:nexora_core/nexora_core.dart';

import '../utils/money.dart';
import '../validators/coupon_validator.dart';

enum CouponDiscountType { fixedMinor, percent }

/// Coupon definition for client-side eligibility checks (server is authoritative).
class CouponDefinition {
  const CouponDefinition({
    required this.code,
    required this.discountType,
    required this.discountValue,
    this.minOrderMinor = 0,
    this.maxDiscountMinor,
    this.expiresAt,
    this.startsAt,
    this.usageLimit,
    this.usageCount = 0,
    this.stackable = false,
    this.firstOrderOnly = false,
    this.currency = 'TRY',
    this.active = true,
    this.applicableCategoryIds = const [],
  });

  final String code;
  final CouponDiscountType discountType;
  final int discountValue;
  final int minOrderMinor;
  final int? maxDiscountMinor;
  final DateTime? expiresAt;
  final DateTime? startsAt;
  final int? usageLimit;
  final int usageCount;
  final bool stackable;
  final bool firstOrderOnly;
  final String currency;
  final bool active;
  final List<String> applicableCategoryIds;

  bool get isExpired =>
      expiresAt != null && !expiresAt!.isAfter(DateTime.now());

  bool get notYetActive =>
      startsAt != null && startsAt!.isAfter(DateTime.now());

  bool get usageLimitReached =>
      usageLimit != null && usageCount >= usageLimit!;
}

enum CouponViolationCode {
  invalidFormat,
  inactive,
  expired,
  notYetActive,
  minOrderNotMet,
  usageLimitReached,
  notStackable,
  firstOrderOnly,
  categoryNotEligible,
}

class CouponViolation {
  const CouponViolation({required this.code, required this.message});

  final CouponViolationCode code;
  final String message;
}

/// Promo stacking, expiry, eligibility, and usage-limit rules.
abstract final class CouponRules {
  static Result<String> validateCodeFormat(String? raw) => CouponValidator.parse(raw);

  static List<CouponViolation> evaluateEligibility({
    required CouponDefinition coupon,
    required int cartSubtotalMinor,
    required String cartCurrency,
    required bool hasActivePromotion,
    required bool isFirstOrder,
    required Set<String> cartCategoryIds,
    DateTime? referenceTime,
  }) {
    final violations = <CouponViolation>[];
    final now = referenceTime ?? DateTime.now();

    if (!coupon.active) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.inactive,
        message: 'This coupon is no longer active',
      ),);
    }

    if (coupon.startsAt != null && coupon.startsAt!.isAfter(now)) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.notYetActive,
        message: 'This coupon is not valid yet',
      ),);
    }

    if (coupon.expiresAt != null && !coupon.expiresAt!.isAfter(now)) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.expired,
        message: 'This coupon has expired',
      ),);
    }

    if (coupon.usageLimit != null && coupon.usageCount >= coupon.usageLimit!) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.usageLimitReached,
        message: 'This coupon has reached its usage limit',
      ),);
    }

    if (coupon.minOrderMinor > 0 && cartSubtotalMinor < coupon.minOrderMinor) {
      final min = Money(minorUnits: coupon.minOrderMinor, currency: coupon.currency);
      violations.add(CouponViolation(
        code: CouponViolationCode.minOrderNotMet,
        message: 'Minimum order ${min.format()} required for this coupon',
      ),);
    }

    if (coupon.currency != cartCurrency) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.inactive,
        message: 'Coupon currency does not match cart',
      ),);
    }

    if (hasActivePromotion && !coupon.stackable) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.notStackable,
        message: 'Coupon cannot be combined with other promotions',
      ),);
    }

    if (coupon.firstOrderOnly && !isFirstOrder) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.firstOrderOnly,
        message: 'Coupon is valid for first orders only',
      ),);
    }

    if (coupon.applicableCategoryIds.isNotEmpty &&
        !coupon.applicableCategoryIds.any(cartCategoryIds.contains)) {
      violations.add(const CouponViolation(
        code: CouponViolationCode.categoryNotEligible,
        message: 'Coupon does not apply to items in your cart',
      ),);
    }

    return violations;
  }

  static Result<CouponDefinition> validateForApply({
    required CouponDefinition coupon,
    required int cartSubtotalMinor,
    required String cartCurrency,
    required bool hasActivePromotion,
    required bool isFirstOrder,
    required Set<String> cartCategoryIds,
    DateTime? referenceTime,
  }) {
    final violations = evaluateEligibility(
      coupon: coupon,
      cartSubtotalMinor: cartSubtotalMinor,
      cartCurrency: cartCurrency,
      hasActivePromotion: hasActivePromotion,
      isFirstOrder: isFirstOrder,
      cartCategoryIds: cartCategoryIds,
      referenceTime: referenceTime,
    );

    if (violations.isNotEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: violations.first.message,
          details: {
            'coupon_code': coupon.code,
            'violations': violations.map((v) => v.code.name).toList(),
          },
        ),
      );
    }

    return Success(coupon);
  }

  static int computeDiscountMinor({
    required CouponDefinition coupon,
    required int cartSubtotalMinor,
  }) {
    final raw = switch (coupon.discountType) {
      CouponDiscountType.fixedMinor => coupon.discountValue,
      CouponDiscountType.percent => (cartSubtotalMinor * coupon.discountValue) ~/ 100,
    };

    if (coupon.maxDiscountMinor != null && raw > coupon.maxDiscountMinor!) {
      return coupon.maxDiscountMinor!;
    }
    return raw.clamp(0, cartSubtotalMinor);
  }

  static bool canStack({
    required CouponDefinition primary,
    required CouponDefinition secondary,
  }) =>
      primary.stackable && secondary.stackable;
}
