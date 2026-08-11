import 'package:nexora_core/nexora_core.dart';

import '../entities/favorites_entity.dart';
import '../repositories/favorites_repository.dart';

class WatchFavoriteEntriesUseCase {
  const WatchFavoriteEntriesUseCase(this._repository);
  final FavoritesRepository _repository;

  Stream<List<FavoriteEntry>> call({FavoriteType? type}) =>
      _repository.watchEntries(type: type);
}

class AddFavoriteEntryUseCase {
  const AddFavoriteEntryUseCase(this._repository);
  final FavoritesRepository _repository;

  Future<Result<FavoriteEntry>> call(FavoriteEntry entry) => _repository.add(entry);
}

class SyncFavoritesUseCase {
  const SyncFavoritesUseCase(this._repository);
  final FavoritesRepository _repository;

  Future<Result<void>> call() => _repository.syncFromCloud();
}
