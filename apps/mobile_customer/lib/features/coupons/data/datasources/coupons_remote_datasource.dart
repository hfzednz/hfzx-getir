import 'package:nexora_core/nexora_core.dart';

import '../../../../shared/business_rules/coupon_rules.dart';
import '../../../../shared/validators/coupon_validator.dart';
import '../../domain/entities/coupons_entity.dart';
import '../models/coupons_model.dart';

class CouponsRemoteDataSource {
  const CouponsRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/coupons';
  static const _validatePath = '/coupons/validate';
  static const _applyPath = '/coupons/apply';

  Future<Result<List<Coupon>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Coupon>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => CouponModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Coupon>> fetchByCode(String code) async {
    return _client.get<Coupon>(
      '$_listPath/$code',
      parser: (json) => CouponModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<CouponApplyResult>> validate({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    Map<String, dynamic>? extra,
  }) async {
    final parsed = CouponValidator.parse(code);
    if (parsed.isFailure) return Failure(parsed.errorOrNull!);

    return _client.post<CouponApplyResult>(
      _validatePath,
      data: {
        'code': parsed.valueOrNull,
        'cart_subtotal_minor': cartSubtotalMinor,
        'cart_currency': cartCurrency,
        ...?extra,
      },
      parser: (json) => CouponApplyResultModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<CouponApplyResult>> apply({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    String? idempotencyKey,
    Map<String, dynamic>? extra,
  }) async {
    final parsed = CouponValidator.parse(code);
    if (parsed.isFailure) return Failure(parsed.errorOrNull!);

    return _client.post<CouponApplyResult>(
      _applyPath,
      data: {
        'code': parsed.valueOrNull,
        'cart_subtotal_minor': cartSubtotalMinor,
        'cart_currency': cartCurrency,
        ...?extra,
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => CouponApplyResultModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  /// Client-side pre-check before hitting the API.
  Result<CouponDefinition> preValidate({
    required Coupon coupon,
    required int cartSubtotalMinor,
    required String cartCurrency,
    required bool hasActivePromotion,
    required bool isFirstOrder,
    required Set<String> cartCategoryIds,
  }) =>
      CouponRules.validateForApply(
        coupon: coupon.toDefinition(),
        cartSubtotalMinor: cartSubtotalMinor,
        cartCurrency: cartCurrency,
        hasActivePromotion: hasActivePromotion,
        isFirstOrder: isFirstOrder,
        cartCategoryIds: cartCategoryIds,
      );
}
