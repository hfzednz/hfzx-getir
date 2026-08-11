import '../../domain/entities/coupons_entity.dart';

class CouponModel {
  const CouponModel({required this.raw});

  final Map<String, dynamic> raw;

  factory CouponModel.fromJson(Map<String, dynamic> json) => CouponModel(raw: json);

  Coupon toEntity() => Coupon.fromJson(raw);
}

class CouponApplyResultModel {
  const CouponApplyResultModel({required this.raw});

  final Map<String, dynamic> raw;

  factory CouponApplyResultModel.fromJson(Map<String, dynamic> json) => CouponApplyResultModel(raw: json);

  CouponApplyResult toEntity() {
    final couponJson = raw['coupon'] as Map<String, dynamic>? ?? raw;
    final stacked = (raw['stacked_coupons'] as List<dynamic>?)
            ?.map((e) => Coupon.fromJson(e as Map<String, dynamic>))
            .toList() ??
        const [];
    return CouponApplyResult(
      coupon: Coupon.fromJson(couponJson),
      discountMinor: (raw['discount_minor'] as num?)?.toInt() ?? 0,
      stackedCoupons: stacked,
      message: raw['message']?.toString() ?? '',
    );
  }
}
