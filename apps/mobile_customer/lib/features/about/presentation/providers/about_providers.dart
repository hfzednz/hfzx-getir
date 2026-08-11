import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/about_remote_datasource.dart';
import '../../data/repositories/about_repository_impl.dart';
import '../../domain/entities/about_entity.dart';
import '../../domain/repositories/about_repository.dart';
import '../../domain/usecases/about_usecases.dart';

final aboutRemoteDataSourceProvider = Provider<AboutRemoteDataSource>((ref) {
  return AboutRemoteDataSource(ref.watch(apiClientProvider));
});

final aboutRepositoryProvider = Provider<AboutRepository>((ref) {
  return AboutRepositoryImpl(ref.watch(aboutRemoteDataSourceProvider));
});

final getAboutUseCaseProvider = Provider((ref) =>
    GetAboutUseCase(ref.watch(aboutRepositoryProvider)));

final listAboutUseCaseProvider = Provider((ref) =>
    ListAboutUseCase(ref.watch(aboutRepositoryProvider)));

final aboutListProvider = FutureProvider.autoDispose<List<AboutInfo>>((ref) async {
  final result = await ref.watch(listAboutUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
