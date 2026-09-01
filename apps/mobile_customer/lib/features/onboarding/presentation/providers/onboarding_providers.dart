import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/onboarding_remote_datasource.dart';
import '../../data/repositories/onboarding_repository_impl.dart';
import '../../domain/entities/onboarding_entity.dart';
import '../../domain/repositories/onboarding_repository.dart';
import '../../domain/usecases/onboarding_usecases.dart';

final onboardingRemoteDataSourceProvider = Provider<OnboardingRemoteDataSource>((ref) {
  return OnboardingRemoteDataSource(ref.watch(apiClientProvider));
});

final onboardingRepositoryProvider = Provider<OnboardingRepository>((ref) {
  return OnboardingRepositoryImpl(ref.watch(onboardingRemoteDataSourceProvider));
});

final getOnboardingUseCaseProvider = Provider((ref) =>
    GetOnboardingUseCase(ref.watch(onboardingRepositoryProvider)),);

final listOnboardingUseCaseProvider = Provider((ref) =>
    ListOnboardingUseCase(ref.watch(onboardingRepositoryProvider)),);

final onboardingListProvider = FutureProvider.autoDispose<List<OnboardingPage>>((ref) async {
  final result = await ref.watch(listOnboardingUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
