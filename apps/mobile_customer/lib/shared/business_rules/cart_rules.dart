import 'package:nexora_core/nexora_core.dart';

import '../../features/cart/domain/entities/cart_entity.dart';
import '../../features/product/domain/entities/product_entity.dart';
import '../utils/money.dart';

/// Per-product metadata used by client-side cart rules (from catalog enrichment).
class CartProductMeta {
  const CartProductMeta({
    this.ageRestricted = false,
    this.weightGrams,
    this.maxQty,
    this.availableQty,
    this.stockStatus = ProductStockStatus.inStock,
  });

  final bool ageRestricted;
  final int? weightGrams;
  final int? maxQty;
  final int? availableQty;
  final ProductStockStatus stockStatus;
}

/// Runtime context for cart rule evaluation.
class CartRulesContext {
  const CartRulesContext({
    this.productMeta = const {},
    this.maxCartWeightGrams = 25000,
    this.userAgeVerified = false,
    this.currency = 'TRY',
  });

  final Map<String, CartProductMeta> productMeta;
  final int maxCartWeightGrams;
  final bool userAgeVerified;
  final String currency;

  CartProductMeta metaFor(String productId) =>
      productMeta[productId] ?? const CartProductMeta();
}

/// Client-side cart business rules applied before server validation.
abstract final class CartRules {
  /// Evaluates all cart violations (backward-compatible entry point).
  static List<CartViolation> evaluate(
    Cart cart, {
    CartRulesContext context = const CartRulesContext(),
  }) {
    return [
      ..._minOrderViolations(cart),
      ..._lineViolations(cart, context),
      ..._weightViolations(cart, context),
      ..._ageViolations(cart, context),
      ..._couponViolations(cart),
    ];
  }

  /// Returns [Success] when the cart passes all rules, otherwise [Failure] with
  /// violation details.
  static Result<Cart> validateForCheckout(
    Cart cart, {
    CartRulesContext context = const CartRulesContext(),
  }) {
    final violations = evaluate(cart, context: context);
    final blocking = violations.where((v) => v.severity == 'error').toList();
    if (blocking.isNotEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: blocking.first.message,
          details: {
            'violations': blocking.map((v) => v.toJson()).toList(),
            'cart_id': cart.id,
          },
        ),
      );
    }
    return Success(_withViolations(cart, violations));
  }

  static bool canProceedToCheckout(
    Cart cart, {
    CartRulesContext context = const CartRulesContext(),
  }) {
    return validateForCheckout(cart, context: context).isSuccess;
  }

  /// Merges anonymous (local) cart with authenticated cloud cart.
  ///
  /// Quantities for matching product+variant keys are summed; cloud pricing wins
  /// when both sides carry a line for the same key.
  static Result<Cart> mergeAnonymousAndCloud({
    required Cart local,
    required Cart cloud,
    CartRulesContext context = const CartRulesContext(),
  }) {
    if (local.items.isEmpty) {
      return validateForCheckout(cloud, context: context);
    }
    if (cloud.items.isEmpty) {
      return validateForCheckout(
        local.copyWith(id: cloud.id.isNotEmpty ? cloud.id : local.id),
        context: context,
      );
    }

    final mergedLines = <String, CartLine>{};

    for (final line in cloud.items) {
      mergedLines[_lineKey(line)] = line;
    }

    for (final line in local.items) {
      final key = _lineKey(line);
      final existing = mergedLines[key];
      if (existing == null) {
        mergedLines[key] = line;
        continue;
      }
      mergedLines[key] = CartLine(
        productId: existing.productId,
        variantId: existing.variantId,
        title: existing.title,
        imageUrl: existing.imageUrl ?? line.imageUrl,
        quantity: existing.quantity + line.quantity,
        unitPriceMinor: existing.unitPriceMinor,
        currency: existing.currency,
        stockStatus: existing.stockStatus,
        stockWarning: existing.stockWarning,
        maxQty: existing.maxQty ?? line.maxQty,
        availableQty: existing.availableQty ?? line.availableQty,
      );
    }

    final items = mergedLines.values.toList();
    final subtotal = items.fold<int>(0, (sum, l) => sum + l.lineTotalMinor);

    final merged = Cart(
      id: cloud.id.isNotEmpty ? cloud.id : local.id,
      items: items,
      promotions: cloud.promotions.isNotEmpty ? cloud.promotions : local.promotions,
      coupon: cloud.coupon ?? local.coupon,
      giftCards: cloud.giftCards.isNotEmpty ? cloud.giftCards : local.giftCards,
      walletAppliedMinor: cloud.walletAppliedMinor,
      loyaltyPointsToRedeem: cloud.loyaltyPointsToRedeem,
      subtotalMinor: subtotal,
      deliveryFeeEstimateMinor: cloud.deliveryFeeEstimateMinor ?? local.deliveryFeeEstimateMinor,
      taxEstimateMinor: cloud.taxEstimateMinor ?? local.taxEstimateMinor,
      totalMinor: cloud.totalMinor > 0 ? cloud.totalMinor : subtotal,
      currency: cloud.currency,
      etaMinutes: cloud.etaMinutes ?? local.etaMinutes,
      minOrderMinor: cloud.minOrderMinor ?? local.minOrderMinor,
    );

    return validateForCheckout(merged, context: context);
  }

  static List<CartViolation> _minOrderViolations(Cart cart) {
    if (cart.items.isEmpty) return const [];
    if (cart.minOrderMinor == null || cart.subtotalMinor >= cart.minOrderMinor!) {
      return const [];
    }
    final money = Money(minorUnits: cart.minOrderMinor!, currency: cart.currency);
    return [
      CartViolation(
        code: CartViolationCode.minOrder,
        message: 'Minimum order is ${money.format()}',
      ),
    ];
  }

  static List<CartViolation> _lineViolations(Cart cart, CartRulesContext context) {
    final violations = <CartViolation>[];
    for (final line in cart.items) {
      final meta = context.metaFor(line.productId);
      final stockStatus = line.stockStatus != ProductStockStatus.inStock
          ? line.stockStatus
          : meta.stockStatus;

      if (stockStatus == ProductStockStatus.outOfStock) {
        violations.add(CartViolation(
          code: CartViolationCode.outOfStock,
          message: '${line.title} is out of stock',
          productId: line.productId,
        ),);
      }

      final maxQty = line.maxQty ?? meta.maxQty;
      if (maxQty != null && line.quantity > maxQty) {
        violations.add(CartViolation(
          code: CartViolationCode.maxQty,
          message: '${line.title} max quantity is $maxQty',
          productId: line.productId,
        ),);
      }

      final availableQty = line.availableQty ?? meta.availableQty;
      if (availableQty != null && line.quantity > availableQty) {
        violations.add(CartViolation(
          code: CartViolationCode.outOfStock,
          message: 'Only $availableQty of ${line.title} available',
          productId: line.productId,
        ),);
      }
    }
    return violations;
  }

  static List<CartViolation> _weightViolations(Cart cart, CartRulesContext context) {
    if (context.maxCartWeightGrams <= 0) return const [];
    var totalGrams = 0;
    for (final line in cart.items) {
      final grams = context.metaFor(line.productId).weightGrams;
      if (grams != null && grams > 0) {
        totalGrams += grams * line.quantity;
      }
    }
    if (totalGrams <= context.maxCartWeightGrams) return const [];
    return [
      CartViolation(
        code: CartViolationCode.maxWeight,
        message:
            'Cart weight exceeds ${context.maxCartWeightGrams ~/ 1000} kg limit',
      ),
    ];
  }

  static List<CartViolation> _ageViolations(Cart cart, CartRulesContext context) {
    if (context.userAgeVerified) return const [];
    for (final line in cart.items) {
      if (context.metaFor(line.productId).ageRestricted) {
        return const [
          CartViolation(
            code: CartViolationCode.ageRestricted,
            message: 'Age verification required for restricted items',
          ),
        ];
      }
    }
    return const [];
  }

  static List<CartViolation> _couponViolations(Cart cart) {
    if (cart.coupon != null && !cart.coupon!.valid) {
      return const [
        CartViolation(
          code: CartViolationCode.couponInvalid,
          message: 'Coupon is not valid',
        ),
      ];
    }
    return const [];
  }

  static Cart _withViolations(Cart cart, List<CartViolation> violations) {
    return Cart(
      id: cart.id,
      items: cart.items,
      promotions: cart.promotions,
      coupon: cart.coupon,
      giftCards: cart.giftCards,
      walletAppliedMinor: cart.walletAppliedMinor,
      loyaltyPointsToRedeem: cart.loyaltyPointsToRedeem,
      subtotalMinor: cart.subtotalMinor,
      deliveryFeeEstimateMinor: cart.deliveryFeeEstimateMinor,
      taxEstimateMinor: cart.taxEstimateMinor,
      totalMinor: cart.totalMinor,
      currency: cart.currency,
      etaMinutes: cart.etaMinutes,
      minOrderMinor: cart.minOrderMinor,
      violations: violations,
    );
  }

  static String _lineKey(CartLine line) => '${line.productId}:${line.variantId ?? ''}';
}

extension on Cart {
  Cart copyWith({String? id}) => Cart(
        id: id ?? this.id,
        items: items,
        promotions: promotions,
        coupon: coupon,
        giftCards: giftCards,
        walletAppliedMinor: walletAppliedMinor,
        loyaltyPointsToRedeem: loyaltyPointsToRedeem,
        subtotalMinor: subtotalMinor,
        deliveryFeeEstimateMinor: deliveryFeeEstimateMinor,
        taxEstimateMinor: taxEstimateMinor,
        totalMinor: totalMinor,
        currency: currency,
        etaMinutes: etaMinutes,
        minOrderMinor: minOrderMinor,
        violations: violations,
      );
}
