import 'package:nexora_core/nexora_core.dart';

import '../entities/coupons_entity.dart';
import '../repositories/coupons_repository.dart';

class ListCouponsUseCase {
  const ListCouponsUseCase(this._repository);
  final CouponsRepository _repository;

  Future<Result<List<Coupon>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class ValidateCouponUseCase {
  const ValidateCouponUseCase(this._repository);
  final CouponsRepository _repository;

  Future<Result<CouponApplyResult>> call({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    Map<String, dynamic>? extra,
  }) =>
      _repository.validate(
        code: code,
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        extra: extra,
      );
}

class ApplyCouponUseCase {
  const ApplyCouponUseCase(this._repository);
  final CouponsRepository _repository;

  Future<Result<CouponApplyResult>> call({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    String? idempotencyKey,
    Map<String, dynamic>? extra,
  }) =>
      _repository.apply(
        code: code,
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        idempotencyKey: idempotencyKey,
        extra: extra,
      );
}
