import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/profile_remote_datasource.dart';
import '../../data/repositories/profile_repository_impl.dart';
import '../../domain/entities/profile_entity.dart';
import '../../domain/repositories/profile_repository.dart';

final profileRemoteDataSourceProvider = Provider<ProfileRemoteDataSource>((ref) {
  return ProfileRemoteDataSource(ref.watch(apiClientProvider));
});

final profileRepositoryProvider = Provider<ProfileRepository>((ref) {
  return ProfileRepositoryImpl(ref.watch(profileRemoteDataSourceProvider));
});

final userProfileProvider = FutureProvider.autoDispose<UserProfile>((ref) async {
  final result = await ref.watch(profileRepositoryProvider).fetchProfile();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final profileUpdateProvider =
    AsyncNotifierProvider<ProfileUpdateController, UserProfile?>(ProfileUpdateController.new);

class ProfileUpdateController extends AsyncNotifier<UserProfile?> {
  @override
  Future<UserProfile?> build() async => null;

  Future<void> save(UserProfile profile) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final result = await ref.read(profileRepositoryProvider).updateProfile(
            profile,
            idempotencyKey: const Uuid().v4(),
          );
      return result.fold(
        onSuccess: (updated) {
          ref.invalidate(userProfileProvider);
          return updated;
        },
        onFailure: (e) => throw e,
      );
    });
  }
}
