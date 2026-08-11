import 'package:equatable/equatable.dart';

import '../../../../shared/business_rules/coupon_rules.dart';

enum CouponStatus { active, expired, used, pending }

class Coupon extends Equatable {
  const Coupon({
    required this.id,
    required this.code,
    this.title = '',
    this.description = '',
    this.discountType = CouponDiscountType.fixedMinor,
    this.discountValue = 0,
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
    this.status = CouponStatus.active,
  });

  final String id;
  final String code;
  final String title;
  final String description;
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
  final CouponStatus status;

  CouponDefinition toDefinition() => CouponDefinition(
        code: code,
        discountType: discountType,
        discountValue: discountValue,
        minOrderMinor: minOrderMinor,
        maxDiscountMinor: maxDiscountMinor,
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

  factory Coupon.fromJson(Map<String, dynamic> json) {
    final discountTypeRaw = json['discount_type']?.toString() ?? json['type']?.toString();
    return Coupon(
      id: json['id']?.toString() ?? json['code']?.toString() ?? '',
      code: (json['code']?.toString() ?? json['id']?.toString() ?? '').toUpperCase(),
      title: json['title']?.toString() ?? json['name']?.toString() ?? '',
      description: json['description']?.toString() ?? '',
      discountType: discountTypeRaw == 'percent' || discountTypeRaw == 'percentage'
          ? CouponDiscountType.percent
          : CouponDiscountType.fixedMinor,
      discountValue: (json['discount_value'] as num?)?.toInt() ??
          (json['value'] as num?)?.toInt() ??
          0,
      minOrderMinor: (json['min_order_minor'] as num?)?.toInt() ?? 0,
      maxDiscountMinor: (json['max_discount_minor'] as num?)?.toInt(),
      expiresAt: json['expires_at'] != null ? DateTime.tryParse(json['expires_at'].toString()) : null,
      startsAt: json['starts_at'] != null ? DateTime.tryParse(json['starts_at'].toString()) : null,
      usageLimit: (json['usage_limit'] as num?)?.toInt(),
      usageCount: (json['usage_count'] as num?)?.toInt() ?? 0,
      stackable: json['stackable'] as bool? ?? false,
      firstOrderOnly: json['first_order_only'] as bool? ?? false,
      currency: json['currency']?.toString() ?? 'TRY',
      active: json['active'] as bool? ?? true,
      applicableCategoryIds: (json['applicable_category_ids'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          const [],
      status: CouponStatus.values.asNameMap()[json['status']?.toString()] ?? CouponStatus.active,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'code': code,
        'title': title,
        'description': description,
        'discount_type': discountType == CouponDiscountType.percent ? 'percent' : 'fixed',
        'discount_value': discountValue,
        'min_order_minor': minOrderMinor,
        if (maxDiscountMinor != null) 'max_discount_minor': maxDiscountMinor,
        if (expiresAt != null) 'expires_at': expiresAt!.toIso8601String(),
        if (startsAt != null) 'starts_at': startsAt!.toIso8601String(),
        if (usageLimit != null) 'usage_limit': usageLimit,
        'usage_count': usageCount,
        'stackable': stackable,
        'first_order_only': firstOrderOnly,
        'currency': currency,
        'active': active,
        'applicable_category_ids': applicableCategoryIds,
        'status': status.name,
      };

  @override
  List<Object?> get props => [id, code, discountType, discountValue, expiresAt, status];
}

class CouponApplyResult extends Equatable {
  const CouponApplyResult({
    required this.coupon,
    required this.discountMinor,
    this.stackedCoupons = const [],
    this.message = '',
  });

  final Coupon coupon;
  final int discountMinor;
  final List<Coupon> stackedCoupons;
  final String message;

  @override
  List<Object?> get props => [coupon, discountMinor, stackedCoupons, message];
}
