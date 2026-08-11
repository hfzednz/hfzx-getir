import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../../shared/business_rules/coupon_rules.dart';
import '../../../../shared/business_rules/coupon_selection_helper.dart';
import '../../data/datasources/coupons_remote_datasource.dart';
import '../../data/repositories/coupons_repository_impl.dart';
import '../../domain/entities/coupons_entity.dart';
import '../../domain/repositories/coupons_repository.dart';
import '../../domain/usecases/coupons_usecases.dart';

final couponsRemoteDataSourceProvider = Provider<CouponsRemoteDataSource>((ref) {
  return CouponsRemoteDataSource(ref.watch(apiClientProvider));
});

final couponsRepositoryProvider = Provider<CouponsRepository>((ref) {
  return CouponsRepositoryImpl(ref.watch(couponsRemoteDataSourceProvider));
});

final listCouponsUseCaseProvider = Provider(
  (ref) => ListCouponsUseCase(ref.watch(couponsRepositoryProvider)),
);

final applyCouponUseCaseProvider = Provider(
  (ref) => ApplyCouponUseCase(ref.watch(couponsRepositoryProvider)),
);

final couponsListProvider = FutureProvider.autoDispose<List<Coupon>>((ref) async {
  final result = await ref.watch(listCouponsUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final couponApplyControllerProvider =
    AsyncNotifierProvider<CouponApplyController, CouponApplyResult?>(CouponApplyController.new);

class CouponApplyController extends AsyncNotifier<CouponApplyResult?> {
  @override
  Future<CouponApplyResult?> build() async => null;

  Future<void> apply({
    required String code,
    required int cartSubtotalMinor,
    required String cartCurrency,
    bool hasActivePromotion = false,
    bool isFirstOrder = false,
    Set<String> cartCategoryIds = const {},
  }) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final listResult = await ref.read(listCouponsUseCaseProvider).call();
      final coupons = listResult.fold(onSuccess: (v) => v, onFailure: (_) => <Coupon>[]);
      final match = coupons.where((c) => c.code == code.toUpperCase()).firstOrNull;

      if (match != null) {
        final pre = CouponRules.validateForApply(
          coupon: match.toDefinition(),
          cartSubtotalMinor: cartSubtotalMinor,
          cartCurrency: cartCurrency,
          hasActivePromotion: hasActivePromotion,
          isFirstOrder: isFirstOrder,
          cartCategoryIds: cartCategoryIds,
        );
        if (pre.isFailure) throw pre.errorOrNull!;
      }

      final result = await ref.read(applyCouponUseCaseProvider).call(
            code: code,
            cartSubtotalMinor: cartSubtotalMinor,
            cartCurrency: cartCurrency,
            idempotencyKey: const Uuid().v4(),
          );

      return result.fold(
        onSuccess: (applied) {
          ref.read(analyticsTrackerProvider).trackRaw(
                eventName: AnalyticsEvents.cartCouponApplied,
                props: {
                  'coupon_code': applied.coupon.code,
                  'discount_minor': applied.discountMinor,
                  'stacked_count': applied.stackedCoupons.length,
                },
              );
          return applied;
        },
        onFailure: (e) => throw e,
      );
    });
  }

  Coupon? suggestBest({
    required List<Coupon> coupons,
    required int cartSubtotalMinor,
    required String cartCurrency,
    bool hasActivePromotion = false,
    bool isFirstOrder = false,
    Set<String> cartCategoryIds = const {},
  }) {
    final best = CouponSelectionHelper.selectBest(
      coupons: coupons.map((c) => c.toDefinition()).toList(),
      cartSubtotalMinor: cartSubtotalMinor,
      cartCurrency: cartCurrency,
      hasActivePromotion: hasActivePromotion,
      isFirstOrder: isFirstOrder,
      cartCategoryIds: cartCategoryIds,
    );
    if (best == null) return null;
    return coupons.firstWhere((c) => c.code == best.code);
  }
}
