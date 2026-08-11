import 'package:equatable/equatable.dart';

import '../../../product/domain/entities/product_entity.dart';

enum CartViolationCode {
  minOrder,
  outOfStock,
  priceChanged,
  couponInvalid,
  ageRestricted,
  maxQty,
  maxWeight,
}

class CartLine extends Equatable {
  const CartLine({
    required this.productId,
    required this.title,
    required this.quantity,
    required this.unitPriceMinor,
    this.variantId,
    this.imageUrl,
    this.currency = 'TRY',
    this.stockStatus = ProductStockStatus.inStock,
    this.stockWarning,
    this.maxQty,
    this.availableQty,
  });

  final String productId;
  final String? variantId;
  final String title;
  final String? imageUrl;
  final int quantity;
  final int unitPriceMinor;
  final String currency;
  final ProductStockStatus stockStatus;
  final String? stockWarning;
  final int? maxQty;
  final int? availableQty;

  int get lineTotalMinor => unitPriceMinor * quantity;

  factory CartLine.fromJson(Map<String, dynamic> json) => CartLine(
        productId: json['product_id']?.toString() ?? json['id']?.toString() ?? '',
        variantId: json['variant_id']?.toString(),
        title: json['title']?.toString() ?? '',
        imageUrl: json['image_url']?.toString(),
        quantity: (json['quantity'] as num?)?.toInt() ?? 1,
        unitPriceMinor: (json['unit_price_minor'] as num?)?.toInt() ??
            (json['price_minor'] as num?)?.toInt() ??
            0,
        currency: json['currency']?.toString() ?? 'TRY',
        stockStatus: productStockStatusFromJson(json['stock_status']?.toString()),
        stockWarning: json['stock_warning']?.toString(),
        maxQty: (json['max_qty'] as num?)?.toInt(),
        availableQty: (json['available_qty'] as num?)?.toInt(),
      );

  Map<String, dynamic> toJson() => {
        'product_id': productId,
        if (variantId != null) 'variant_id': variantId,
        'title': title,
        if (imageUrl != null) 'image_url': imageUrl,
        'quantity': quantity,
        'unit_price_minor': unitPriceMinor,
        'currency': currency,
        'stock_status': productStockStatusToJson(stockStatus),
        if (stockWarning != null) 'stock_warning': stockWarning,
        if (maxQty != null) 'max_qty': maxQty,
        if (availableQty != null) 'available_qty': availableQty,
      };

  @override
  List<Object?> get props => [
        productId,
        variantId,
        title,
        imageUrl,
        quantity,
        unitPriceMinor,
        currency,
        stockStatus,
        stockWarning,
        maxQty,
        availableQty,
      ];
}

class CartPromotion extends Equatable {
  const CartPromotion({
    required this.id,
    required this.title,
    required this.discountMinor,
    this.description,
    this.code,
  });

  final String id;
  final String title;
  final int discountMinor;
  final String? description;
  final String? code;

  factory CartPromotion.fromJson(Map<String, dynamic> json) => CartPromotion(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        discountMinor: (json['discount_minor'] as num?)?.toInt() ?? 0,
        description: json['description']?.toString(),
        code: json['code']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        'discount_minor': discountMinor,
        if (description != null) 'description': description,
        if (code != null) 'code': code,
      };

  @override
  List<Object?> get props => [id, title, discountMinor, description, code];
}

class CartCoupon extends Equatable {
  const CartCoupon({
    required this.code,
    required this.discountMinor,
    this.title,
    this.valid = true,
  });

  final String code;
  final int discountMinor;
  final String? title;
  final bool valid;

  factory CartCoupon.fromJson(Map<String, dynamic> json) => CartCoupon(
        code: json['code']?.toString() ?? '',
        discountMinor: (json['discount_minor'] as num?)?.toInt() ?? 0,
        title: json['title']?.toString(),
        valid: json['valid'] != false,
      );

  Map<String, dynamic> toJson() => {
        'code': code,
        'discount_minor': discountMinor,
        if (title != null) 'title': title,
        'valid': valid,
      };

  @override
  List<Object?> get props => [code, discountMinor, title, valid];
}

class CartGiftCard extends Equatable {
  const CartGiftCard({
    required this.code,
    required this.appliedMinor,
    this.balanceMinor,
  });

  final String code;
  final int appliedMinor;
  final int? balanceMinor;

  factory CartGiftCard.fromJson(Map<String, dynamic> json) => CartGiftCard(
        code: json['code']?.toString() ?? '',
        appliedMinor: (json['applied_minor'] as num?)?.toInt() ?? 0,
        balanceMinor: (json['balance_minor'] as num?)?.toInt(),
      );

  Map<String, dynamic> toJson() => {
        'code': code,
        'applied_minor': appliedMinor,
        if (balanceMinor != null) 'balance_minor': balanceMinor,
      };

  @override
  List<Object?> get props => [code, appliedMinor, balanceMinor];
}

class CartViolation extends Equatable {
  const CartViolation({
    required this.code,
    required this.message,
    this.productId,
    this.severity = 'error',
  });

  final CartViolationCode code;
  final String message;
  final String? productId;
  final String severity;

  factory CartViolation.fromJson(Map<String, dynamic> json) {
    CartViolationCode parseCode(String? raw) {
      switch (raw?.toLowerCase()) {
        case 'min_order':
          return CartViolationCode.minOrder;
        case 'out_of_stock':
          return CartViolationCode.outOfStock;
        case 'price_changed':
          return CartViolationCode.priceChanged;
        case 'coupon_invalid':
          return CartViolationCode.couponInvalid;
        case 'age_restricted':
          return CartViolationCode.ageRestricted;
        case 'max_qty':
          return CartViolationCode.maxQty;
        case 'max_weight':
          return CartViolationCode.maxWeight;
        default:
          return CartViolationCode.minOrder;
      }
    }

    return CartViolation(
      code: parseCode(json['code']?.toString()),
      message: json['message']?.toString() ?? '',
      productId: json['product_id']?.toString(),
      severity: json['severity']?.toString() ?? 'error',
    );
  }

  Map<String, dynamic> toJson() => {
        'code': code.name,
        'message': message,
        if (productId != null) 'product_id': productId,
        'severity': severity,
      };

  @override
  List<Object?> get props => [code, message, productId, severity];
}

class Cart extends Equatable {
  const Cart({
    required this.id,
    this.items = const [],
    this.promotions = const [],
    this.coupon,
    this.giftCards = const [],
    this.walletAppliedMinor = 0,
    this.loyaltyPointsToRedeem = 0,
    this.subtotalMinor = 0,
    this.deliveryFeeEstimateMinor,
    this.taxEstimateMinor,
    this.totalMinor = 0,
    this.currency = 'TRY',
    this.etaMinutes,
    this.minOrderMinor,
    this.violations = const [],
  });

  final String id;
  final List<CartLine> items;
  final List<CartPromotion> promotions;
  final CartCoupon? coupon;
  final List<CartGiftCard> giftCards;
  final int walletAppliedMinor;
  final int loyaltyPointsToRedeem;
  final int subtotalMinor;
  final int? deliveryFeeEstimateMinor;
  final int? taxEstimateMinor;
  final int totalMinor;
  final String currency;
  final int? etaMinutes;
  final int? minOrderMinor;
  final List<CartViolation> violations;

  bool get hasViolations => violations.isNotEmpty;
  bool get canCheckout => violations.where((v) => v.severity == 'error').isEmpty;

  factory Cart.fromJson(Map<String, dynamic> json) => Cart(
        id: json['id']?.toString() ?? '',
        items: (json['items'] as List<dynamic>? ?? [])
            .map((e) => CartLine.fromJson(e as Map<String, dynamic>))
            .toList(),
        promotions: (json['promotions'] as List<dynamic>? ?? [])
            .map((e) => CartPromotion.fromJson(e as Map<String, dynamic>))
            .toList(),
        coupon: json['coupon'] != null
            ? CartCoupon.fromJson(json['coupon'] as Map<String, dynamic>)
            : null,
        giftCards: (json['gift_cards'] as List<dynamic>? ?? [])
            .map((e) => CartGiftCard.fromJson(e as Map<String, dynamic>))
            .toList(),
        walletAppliedMinor: (json['wallet_applied_minor'] as num?)?.toInt() ?? 0,
        loyaltyPointsToRedeem: (json['loyalty_points_to_redeem'] as num?)?.toInt() ?? 0,
        subtotalMinor: (json['subtotal_minor'] as num?)?.toInt() ?? 0,
        deliveryFeeEstimateMinor: (json['delivery_fee_estimate_minor'] as num?)?.toInt(),
        taxEstimateMinor: (json['tax_estimate_minor'] as num?)?.toInt(),
        totalMinor: (json['total_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
        etaMinutes: (json['eta_minutes'] as num?)?.toInt(),
        minOrderMinor: (json['min_order_minor'] as num?)?.toInt(),
        violations: (json['violations'] as List<dynamic>? ?? [])
            .map((e) => CartViolation.fromJson(e as Map<String, dynamic>))
            .toList(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'items': items.map((e) => e.toJson()).toList(),
        'promotions': promotions.map((e) => e.toJson()).toList(),
        if (coupon != null) 'coupon': coupon!.toJson(),
        'gift_cards': giftCards.map((e) => e.toJson()).toList(),
        'wallet_applied_minor': walletAppliedMinor,
        'loyalty_points_to_redeem': loyaltyPointsToRedeem,
        'subtotal_minor': subtotalMinor,
        if (deliveryFeeEstimateMinor != null)
          'delivery_fee_estimate_minor': deliveryFeeEstimateMinor,
        if (taxEstimateMinor != null) 'tax_estimate_minor': taxEstimateMinor,
        'total_minor': totalMinor,
        'currency': currency,
        if (etaMinutes != null) 'eta_minutes': etaMinutes,
        if (minOrderMinor != null) 'min_order_minor': minOrderMinor,
        'violations': violations.map((e) => e.toJson()).toList(),
      };

  @override
  List<Object?> get props => [
        id,
        items,
        promotions,
        coupon,
        giftCards,
        walletAppliedMinor,
        loyaltyPointsToRedeem,
        subtotalMinor,
        deliveryFeeEstimateMinor,
        taxEstimateMinor,
        totalMinor,
        currency,
        etaMinutes,
        minOrderMinor,
        violations,
      ];
}
