import 'dart:io';

import 'package:drift/drift.dart';
import 'package:drift/native.dart';
import 'package:path/path.dart' as p;
import 'package:path_provider/path_provider.dart';

part 'app_database.g.dart';

class CartItems extends Table {
  TextColumn get productId => text()();
  TextColumn get variantId => text().nullable()();
  TextColumn get title => text()();
  TextColumn get imageUrl => text().nullable()();
  IntColumn get quantity => integer().withDefault(const Constant(1))();
  IntColumn get unitPriceMinor => integer()();
  TextColumn get currency => text().withDefault(const Constant('TRY'))();
  TextColumn get notes => text().nullable()();
  DateTimeColumn get updatedAt => dateTime().withDefault(currentDateAndTime)();
  BoolColumn get pendingSync => boolean().withDefault(const Constant(false))();

  @override
  Set<Column> get primaryKey => {productId, variantId};
}

class Favorites extends Table {
  TextColumn get productId => text()();
  TextColumn get title => text()();
  TextColumn get imageUrl => text().nullable()();
  IntColumn get unitPriceMinor => integer().nullable()();
  DateTimeColumn get addedAt => dateTime().withDefault(currentDateAndTime)();

  @override
  Set<Column> get primaryKey => {productId};
}

@DataClassName('FavoriteEntryRow')
class FavoriteEntries extends Table {
  TextColumn get entryId => text()();
  TextColumn get type => text()();
  TextColumn get targetId => text()();
  TextColumn get title => text()();
  TextColumn get subtitle => text().withDefault(const Constant(''))();
  TextColumn get imageUrl => text().nullable()();
  TextColumn get metadataJson => text().withDefault(const Constant('{}'))();
  DateTimeColumn get addedAt => dateTime().withDefault(currentDateAndTime)();
  BoolColumn get pendingSync => boolean().withDefault(const Constant(false))();

  @override
  Set<Column> get primaryKey => {entryId};
}

class RecentSearches extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get query => text()();
  DateTimeColumn get searchedAt => dateTime().withDefault(currentDateAndTime)();
}

class CachedOrders extends Table {
  TextColumn get orderId => text()();
  TextColumn get status => text()();
  TextColumn get payloadJson => text()();
  DateTimeColumn get cachedAt => dateTime().withDefault(currentDateAndTime)();

  @override
  Set<Column> get primaryKey => {orderId};
}

class SyncQueue extends Table {
  TextColumn get clientMutationId => text()();
  TextColumn get method => text()();
  TextColumn get path => text()();
  TextColumn get bodyJson => text().nullable()();
  TextColumn get idempotencyKey => text()();
  IntColumn get retryCount => integer().withDefault(const Constant(0))();
  DateTimeColumn get createdAt => dateTime().withDefault(currentDateAndTime)();
  DateTimeColumn get nextAttemptAt => dateTime().nullable()();
  TextColumn get lastError => text().nullable()();

  @override
  Set<Column> get primaryKey => {clientMutationId};
}

@DriftDatabase(tables: [CartItems, Favorites, FavoriteEntries, RecentSearches, CachedOrders, SyncQueue])
class AppDatabase extends _$AppDatabase {
  AppDatabase([QueryExecutor? executor]) : super(executor ?? _openConnection());

  @override
  int get schemaVersion => 2;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onCreate: (m) async => m.createAll(),
        onUpgrade: (m, from, to) async {
          if (from < 2) {
            await m.createTable(favoriteEntries);
          }
        },
      );

  static LazyDatabase _openConnection() {
    return LazyDatabase(() async {
      final dir = await getApplicationDocumentsDirectory();
      final file = File(p.join(dir.path, 'nexora_customer.sqlite'));
      return NativeDatabase.createInBackground(file);
    });
  }

  Stream<List<CartItem>> watchCartItems() => select(cartItems).watch();

  Future<void> upsertCartItem(CartItemsCompanion item) =>
      into(cartItems).insertOnConflictUpdate(item);

  Future<void> removeCartItem(String productId, {String? variantId}) =>
      (delete(cartItems)..where(
            (t) =>
                t.productId.equals(productId) &
                (variantId == null
                    ? t.variantId.isNull()
                    : t.variantId.equals(variantId)),
          ))
          .go();

  Stream<List<Favorite>> watchFavorites() => select(favorites).watch();

  Stream<List<FavoriteEntryRow>> watchFavoriteEntries({String? type}) {
    final q = select(favoriteEntries);
    if (type != null) {
      q.where((t) => t.type.equals(type));
    }
    q.orderBy([(t) => OrderingTerm.desc(t.addedAt)]);
    return q.watch();
  }

  Future<void> upsertFavoriteEntry(FavoriteEntriesCompanion entry) =>
      into(favoriteEntries).insertOnConflictUpdate(entry);

  Future<void> removeFavoriteEntry(String entryId) =>
      (delete(favoriteEntries)..where((t) => t.entryId.equals(entryId))).go();

  Future<List<FavoriteEntryRow>> favoriteEntriesPendingSync() =>
      (select(favoriteEntries)..where((t) => t.pendingSync.equals(true))).get();

  Future<void> addRecentSearch(String query) async {
    await into(recentSearches).insert(
      RecentSearchesCompanion.insert(query: query),
    );
    final all = await select(recentSearches).get();
    if (all.length > 20) {
      all.sort((a, b) => b.searchedAt.compareTo(a.searchedAt));
      for (final old in all.skip(20)) {
        await (delete(recentSearches)..where((t) => t.id.equals(old.id))).go();
      }
    }
  }

  Future<List<RecentSearche>> getRecentSearches({int limit = 10}) async {
    final q = select(recentSearches)
      ..orderBy([(t) => OrderingTerm.desc(t.searchedAt)])
      ..limit(limit);
    return q.get();
  }

  Stream<List<RecentSearche>> watchRecentSearches({int limit = 10}) {
    final q = select(recentSearches)
      ..orderBy([(t) => OrderingTerm.desc(t.searchedAt)])
      ..limit(limit);
    return q.watch();
  }

  Future<void> clearRecentSearches() => delete(recentSearches).go();

  Future<void> cacheOrder(String orderId, String status, String payloadJson) =>
      into(cachedOrders).insertOnConflictUpdate(
        CachedOrdersCompanion.insert(
          orderId: orderId,
          status: status,
          payloadJson: payloadJson,
        ),
      );

  Stream<List<CachedOrder>> watchCachedOrders() {
    final q = select(cachedOrders)
      ..orderBy([(t) => OrderingTerm.desc(t.cachedAt)]);
    return q.watch();
  }
}
