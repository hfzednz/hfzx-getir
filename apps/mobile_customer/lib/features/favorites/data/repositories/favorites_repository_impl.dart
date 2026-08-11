import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../domain/entities/favorites_entity.dart';
import '../../domain/repositories/favorites_repository.dart';
import '../datasources/favorites_local_datasource.dart';
import '../datasources/favorites_remote_datasource.dart';

class FavoritesRepositoryImpl implements FavoritesRepository {
  FavoritesRepositoryImpl(this._local, this._remote);

  final FavoritesLocalDataSource _local;
  final FavoritesRemoteDataSource _remote;

  @override
  Stream<List<FavoriteEntry>> watchEntries({FavoriteType? type}) =>
      _local.watchEntries(type: type);

  @override
  Future<Result<FavoriteEntry>> add(FavoriteEntry entry) async {
    final localEntry = FavoriteEntry(
      id: entry.id.isNotEmpty ? entry.id : '${entry.type.name}:${entry.targetId}',
      type: entry.type,
      targetId: entry.targetId,
      title: entry.title,
      subtitle: entry.subtitle,
      imageUrl: entry.imageUrl,
      metadata: entry.metadata,
      addedAt: entry.addedAt ?? DateTime.now(),
      pendingSync: true,
    );
    await _local.upsert(localEntry);

    final remote = await _remote.add(localEntry, idempotencyKey: const Uuid().v4());
    return remote.fold(
      onSuccess: (synced) async {
        await _local.upsert(synced.copyWith(pendingSync: false));
        return Success(synced);
      },
      onFailure: (e) => Success(localEntry),
    );
  }

  @override
  Future<Result<void>> remove(String entryId) async {
    await _local.remove(entryId);
    return _remote.remove(entryId);
  }

  @override
  Future<Result<void>> syncFromCloud() async {
    final remote = await _remote.fetchList();
    return remote.fold(
      onSuccess: (entries) async {
        for (final e in entries) {
          await _local.upsert(e.copyWith(pendingSync: false));
        }
        final pending = await _local.pendingSync();
        if (pending.isNotEmpty) {
          await _remote.syncBatch(pending);
          for (final e in pending) {
            await _local.upsert(e.copyWith(pendingSync: false));
          }
        }
        return const Success(null);
      },
      onFailure: (e) => Failure(e),
    );
  }
}
