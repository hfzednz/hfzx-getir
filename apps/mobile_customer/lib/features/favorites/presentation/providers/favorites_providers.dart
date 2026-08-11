import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/favorites_local_datasource.dart';
import '../../data/datasources/favorites_remote_datasource.dart';
import '../../data/repositories/favorites_repository_impl.dart';
import '../../domain/entities/favorites_entity.dart';
import '../../domain/repositories/favorites_repository.dart';

final favoritesLocalDataSourceProvider = Provider<FavoritesLocalDataSource>((ref) {
  return FavoritesLocalDataSource(ref.watch(databaseProvider));
});

final favoritesRemoteDataSourceProvider = Provider<FavoritesRemoteDataSource>((ref) {
  return FavoritesRemoteDataSource(ref.watch(apiClientProvider));
});

final favoritesRepositoryProvider = Provider<FavoritesRepository>((ref) {
  return FavoritesRepositoryImpl(
    ref.watch(favoritesLocalDataSourceProvider),
    ref.watch(favoritesRemoteDataSourceProvider),
  );
});

final favoriteEntriesStreamProvider =
    StreamProvider.autoDispose.family<List<FavoriteEntry>, FavoriteType?>((ref, type) {
  return ref.watch(favoritesRepositoryProvider).watchEntries(type: type);
});

final favoritesSyncProvider = FutureProvider.autoDispose<void>((ref) async {
  final result = await ref.watch(favoritesRepositoryProvider).syncFromCloud();
  result.fold(onSuccess: (_) {}, onFailure: (e) => throw e);
});

final favoriteTypeFilterProvider = StateProvider<FavoriteType?>((ref) => null);
