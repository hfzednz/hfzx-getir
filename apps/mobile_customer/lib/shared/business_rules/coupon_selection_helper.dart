import 'coupon_rules.dart';

/// Picks the coupon that yields the highest discount for the current cart context.
abstract final class CouponSelectionHelper {
  static CouponDefinition? selectBest({
    required List<CouponDefinition> coupons,
    required int cartSubtotalMinor,
    required String cartCurrency,
    required bool hasActivePromotion,
    required bool isFirstOrder,
    required Set<String> cartCategoryIds,
    DateTime? referenceTime,
  }) {
    CouponDefinition? best;
    var bestDiscount = -1;

    for (final coupon in coupons) {
      final violations = CouponRules.evaluateEligibility(
        coupon: coupon,
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        hasActivePromotion: hasActivePromotion,
        isFirstOrder: isFirstOrder,
        cartCategoryIds: cartCategoryIds,
        referenceTime: referenceTime,
      );
      if (violations.isNotEmpty) continue;

      final discount = CouponRules.computeDiscountMinor(
        coupon: coupon,
        cartSubtotalMinor: cartSubtotalMinor,
      );
      if (discount > bestDiscount) {
        bestDiscount = discount;
        best = coupon;
      }
    }

    return best;
  }

  static List<CouponDefinition> stackableGroup({
    required List<CouponDefinition> selected,
    required CouponDefinition candidate,
  }) {
    if (selected.isEmpty) return [candidate];
    if (selected.every((c) => CouponRules.canStack(primary: c, secondary: candidate))) {
      return [...selected, candidate];
    }
    return selected;
  }
}
