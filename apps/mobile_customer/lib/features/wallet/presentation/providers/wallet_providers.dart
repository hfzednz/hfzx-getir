import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../data/datasources/wallet_remote_datasource.dart';
import '../../data/repositories/wallet_repository_impl.dart';
import '../../domain/entities/wallet_entity.dart';
import '../../domain/repositories/wallet_repository.dart';
import '../../domain/usecases/wallet_usecases.dart';

final walletRemoteDataSourceProvider = Provider<WalletRemoteDataSource>((ref) {
  return WalletRemoteDataSource(ref.watch(apiClientProvider));
});

final walletRepositoryProvider = Provider<WalletRepository>((ref) {
  return WalletRepositoryImpl(ref.watch(walletRemoteDataSourceProvider));
});

final getWalletAccountUseCaseProvider = Provider(
  (ref) => GetWalletAccountUseCase(ref.watch(walletRepositoryProvider)),
);

final listWalletTransactionsUseCaseProvider = Provider(
  (ref) => ListWalletTransactionsUseCase(ref.watch(walletRepositoryProvider)),
);

final topUpWalletUseCaseProvider = Provider(
  (ref) => TopUpWalletUseCase(ref.watch(walletRepositoryProvider)),
);

final walletAccountProvider = FutureProvider.autoDispose<WalletAccount>((ref) async {
  final result = await ref.watch(getWalletAccountUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final walletTransactionsProvider =
    FutureProvider.autoDispose<List<WalletTransaction>>((ref) async {
  final result = await ref.watch(listWalletTransactionsUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final walletTopUpControllerProvider =
    AsyncNotifierProvider<WalletTopUpController, void>(WalletTopUpController.new);

class WalletTopUpController extends AsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<void> topUp({required int amountMinor, String currency = 'TRY'}) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final result = await ref.read(topUpWalletUseCaseProvider).call(
            amountMinor: amountMinor,
            currency: currency,
            idempotencyKey: const Uuid().v4(),
          );
      result.fold(
        onSuccess: (_) {
          ref.invalidate(walletAccountProvider);
          ref.invalidate(walletTransactionsProvider);
          ref.read(analyticsTrackerProvider).trackRaw(
                eventName: AnalyticsEvents.walletTopUpSucceeded,
                props: {'amount_minor': amountMinor, 'currency': currency},
              );
        },
        onFailure: (e) => throw e,
      );
    });
  }
}
