import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/profile_remote_datasource.dart';
import '../../data/repositories/profile_repository_impl.dart';
import '../../domain/entities/profile_entity.dart';
import '../../domain/repositories/profile_repository.dart';

final profileRemoteDataSourceProvider =
    Provider<ProfileRemoteDataSource>((ref) {
  return ProfileRemoteDataSource(ref.watch(apiClientProvider));
});

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  return ProfileRepositoryImpl(ref.watch(profileRemoteDataSourceProvider));
});

final profileProvider =
    FutureProvider.autoDispose<CourierProfile>((ref) async {
  final result = await ref.watch(profileRepositoryProvider).fetch();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
