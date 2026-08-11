import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/coupons_entity.dart';
import '../../domain/repositories/coupons_repository.dart';
import '../datasources/coupons_remote_datasource.dart';

class CouponsRepositoryImpl implements CouponsRepository {
  const CouponsRepositoryImpl(this._remote);
  final CouponsRemoteDataSource _remote;

  @override
  Future<Result<List<Coupon>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Coupon>> fetchByCode(String code) => _remote.fetchByCode(code);

  @override
  Future<Result<CouponApplyResult>> validate({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    Map<String, dynamic>? extra,
  }) =>
      _remote.validate(
        code: code,
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        extra: extra,
      );

  @override
  Future<Result<CouponApplyResult>> apply({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    String? idempotencyKey,
    Map<String, dynamic>? extra,
  }) =>
      _remote.apply(
        code: code,
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        idempotencyKey: idempotencyKey,
        extra: extra,
      );
}
