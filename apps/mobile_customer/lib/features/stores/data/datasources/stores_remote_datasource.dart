import 'package:nexora_core/nexora_core.dart';

import '../../../../shared/utils/json_list.dart';
import '../../domain/entities/store_entity.dart';

class StoresRemoteDataSource {
  const StoresRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<StoreSummary>>> fetchList() {
    return _client.get<List<StoreSummary>>(
      '/stores',
      parser: (json) =>
          jsonObjectList(json).map(StoreSummary.fromJson).toList(),
    );
  }

  Future<Result<StoreSummary>> fetch(String id) {
    return _client.get<StoreSummary>(
      '/stores/$id',
      parser: (json) =>
          StoreSummary.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
