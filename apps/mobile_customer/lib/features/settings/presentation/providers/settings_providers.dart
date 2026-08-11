import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/settings_remote_datasource.dart';
import '../../data/repositories/settings_repository_impl.dart';
import '../../domain/entities/settings_entity.dart';
import '../../domain/repositories/settings_repository.dart';
import '../../domain/usecases/settings_usecases.dart';

final settingsRemoteDataSourceProvider = Provider<SettingsRemoteDataSource>((ref) {
  return SettingsRemoteDataSource(ref.watch(apiClientProvider));
});

final settingsRepositoryProvider = Provider<SettingsRepository>((ref) {
  return SettingsRepositoryImpl(ref.watch(settingsRemoteDataSourceProvider));
});

final getSettingsUseCaseProvider = Provider((ref) =>
    GetSettingsUseCase(ref.watch(settingsRepositoryProvider)));

final listSettingsUseCaseProvider = Provider((ref) =>
    ListSettingsUseCase(ref.watch(settingsRepositoryProvider)));

final settingsListProvider = FutureProvider.autoDispose<List<AppSettings>>((ref) async {
  final result = await ref.watch(listSettingsUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
