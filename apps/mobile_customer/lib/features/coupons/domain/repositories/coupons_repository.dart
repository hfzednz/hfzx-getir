import 'package:nexora_core/nexora_core.dart';

import '../entities/coupons_entity.dart';

abstract class CouponsRepository {
  Future<Result<List<Coupon>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Coupon>> fetchByCode(String code);
  Future<Result<CouponApplyResult>> validate({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    Map<String, dynamic>? extra,
  });
  Future<Result<CouponApplyResult>> apply({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    String? idempotencyKey,
    Map<String, dynamic>? extra,
  });
}
