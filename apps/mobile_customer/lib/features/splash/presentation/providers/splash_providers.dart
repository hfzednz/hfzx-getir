import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/splash_remote_datasource.dart';
import '../../data/repositories/splash_repository_impl.dart';
import '../../domain/entities/splash_entity.dart';
import '../../domain/repositories/splash_repository.dart';
import '../../domain/usecases/splash_usecases.dart';

final splashRemoteDataSourceProvider = Provider<SplashRemoteDataSource>((ref) {
  return SplashRemoteDataSource(ref.watch(apiClientProvider));
});

final splashRepositoryProvider = Provider<SplashRepository>((ref) {
  return SplashRepositoryImpl(ref.watch(splashRemoteDataSourceProvider));
});

final getSplashUseCaseProvider = Provider((ref) =>
    GetSplashUseCase(ref.watch(splashRepositoryProvider)));

final listSplashUseCaseProvider = Provider((ref) =>
    ListSplashUseCase(ref.watch(splashRepositoryProvider)));

final splashListProvider = FutureProvider.autoDispose<List<SplashState>>((ref) async {
  final result = await ref.watch(listSplashUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
