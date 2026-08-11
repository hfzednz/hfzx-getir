import 'package:nexora_core/nexora_core.dart';

import '../entities/favorites_entity.dart';

abstract class FavoritesRepository {
  Stream<List<FavoriteEntry>> watchEntries({FavoriteType? type});
  Future<Result<FavoriteEntry>> add(FavoriteEntry entry);
  Future<Result<void>> remove(String entryId);
  Future<Result<void>> syncFromCloud();
}
