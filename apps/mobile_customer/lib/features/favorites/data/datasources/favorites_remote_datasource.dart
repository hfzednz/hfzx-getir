import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/favorites_entity.dart';
import '../models/favorites_model.dart';

class FavoritesRemoteDataSource {
  const FavoritesRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/favorites';
  static const _mutatePath = '/favorites';

  Future<Result<List<FavoriteEntry>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<FavoriteEntry>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => FavoriteEntryModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<FavoriteEntry>> add(FavoriteEntry entry, {String? idempotencyKey}) async {
    return _client.post<FavoriteEntry>(
      _mutatePath,
      data: entry.toJson(),
      idempotencyKey: idempotencyKey,
      parser: (json) => FavoriteEntryModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<void>> remove(String entryId) async {
    return _client.delete<void>('$_mutatePath/$entryId', parser: (_) {});
  }

  Future<Result<void>> syncBatch(List<FavoriteEntry> entries) async {
    return _client.post<void>(
      '$_listPath/sync',
      data: {'entries': entries.map((e) => e.toJson()).toList()},
      parser: (_) {},
    );
  }
}
