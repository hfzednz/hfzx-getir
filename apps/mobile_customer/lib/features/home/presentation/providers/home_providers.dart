import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../domain/entities/home_entity.dart';
import '../../data/datasources/home_remote_datasource.dart';
import '../../data/repositories/home_repository_impl.dart';
import '../../domain/repositories/home_repository.dart';
import '../../domain/usecases/home_usecases.dart';

final homeRemoteDataSourceProvider = Provider<HomeRemoteDataSource>((ref) {
  return HomeRemoteDataSource(ref.watch(apiClientProvider));
});

final homeRepositoryProvider = Provider<HomeRepository>((ref) {
  return HomeRepositoryImpl(ref.watch(homeRemoteDataSourceProvider));
});

final getHomeFeedUseCaseProvider = Provider(
  (ref) => GetHomeUseCase(ref.watch(homeRepositoryProvider)),
);

final homeFeedProvider = FutureProvider.autoDispose<HomeFeed>((ref) async {
  final result = await ref.watch(getHomeFeedUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
