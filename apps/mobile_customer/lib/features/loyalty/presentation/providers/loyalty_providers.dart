import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/loyalty_remote_datasource.dart';
import '../../data/repositories/loyalty_repository_impl.dart';
import '../../domain/entities/loyalty_entity.dart';
import '../../domain/repositories/loyalty_repository.dart';
import '../../domain/usecases/loyalty_usecases.dart';

final loyaltyRemoteDataSourceProvider = Provider<LoyaltyRemoteDataSource>((ref) {
  return LoyaltyRemoteDataSource(ref.watch(apiClientProvider));
});

final loyaltyRepositoryProvider = Provider<LoyaltyRepository>((ref) {
  return LoyaltyRepositoryImpl(ref.watch(loyaltyRemoteDataSourceProvider));
});

final loyaltyAccountProvider = FutureProvider.autoDispose<LoyaltyAccount>((ref) async {
  final result = await GetLoyaltyAccountUseCase(ref.watch(loyaltyRepositoryProvider)).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final loyaltyAchievementsProvider =
    FutureProvider.autoDispose<List<LoyaltyAchievement>>((ref) async {
  final result = await ListLoyaltyAchievementsUseCase(ref.watch(loyaltyRepositoryProvider)).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final loyaltyBadgesProvider = FutureProvider.autoDispose<List<LoyaltyBadge>>((ref) async {
  final result = await ref.watch(loyaltyRepositoryProvider).fetchBadges();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final loyaltyMilestonesProvider = FutureProvider.autoDispose<List<LoyaltyMilestone>>((ref) async {
  final result = await ref.watch(loyaltyRepositoryProvider).fetchMilestones();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final dailyRewardClaimProvider =
    AsyncNotifierProvider<DailyRewardClaimController, void>(DailyRewardClaimController.new);

class DailyRewardClaimController extends AsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<void> claim() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final result = await ClaimDailyRewardUseCase(ref.read(loyaltyRepositoryProvider)).call(
            idempotencyKey: const Uuid().v4(),
          );
      result.fold(
        onSuccess: (_) {
          ref.invalidate(loyaltyAccountProvider);
          ref.invalidate(loyaltyAchievementsProvider);
        },
        onFailure: (e) => throw e,
      );
    });
  }
}
