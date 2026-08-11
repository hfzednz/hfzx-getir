import 'dart:convert';

import 'package:drift/drift.dart';

import '../../../cart/data/local/app_database.dart';
import '../../domain/entities/favorites_entity.dart' as domain;

class FavoritesLocalDataSource {
  const FavoritesLocalDataSource(this._db);
  final AppDatabase _db;

  Stream<List<domain.FavoriteEntry>> watchEntries({domain.FavoriteType? type}) =>
      _db.watchFavoriteEntries(type: type?.name).map((rows) => rows.map(_rowToEntity).toList());

  Future<void> upsert(domain.FavoriteEntry entry) => _db.upsertFavoriteEntry(
        FavoriteEntriesCompanion.insert(
          entryId: entry.id,
          type: entry.type.name,
          targetId: entry.targetId,
          title: entry.title,
          subtitle: Value(entry.subtitle),
          imageUrl: Value(entry.imageUrl),
          metadataJson: Value(jsonEncode(entry.metadata)),
          addedAt: Value(entry.addedAt ?? DateTime.now()),
          pendingSync: Value(entry.pendingSync),
        ),
      );

  Future<void> remove(String entryId) => _db.removeFavoriteEntry(entryId);

  Future<List<domain.FavoriteEntry>> pendingSync() async {
    final rows = await _db.favoriteEntriesPendingSync();
    return rows.map(_rowToEntity).toList();
  }

  domain.FavoriteEntry _rowToEntity(FavoriteEntryRow row) => domain.FavoriteEntry(
        id: row.entryId,
        type: domain.FavoriteType.values.asNameMap()[row.type] ?? domain.FavoriteType.product,
        targetId: row.targetId,
        title: row.title,
        subtitle: row.subtitle,
        imageUrl: row.imageUrl,
        metadata: Map<String, dynamic>.from(jsonDecode(row.metadataJson) as Map),
        addedAt: row.addedAt,
        pendingSync: row.pendingSync,
      );
}
