import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/settings_entity.dart';
import '../models/settings_model.dart';

class SettingsRemoteDataSource {
  const SettingsRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/settings';
  static const _mutatePath = '/settings';

  Future<Result<AppSettings>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<AppSettings>(
      path,
      parser: (json) => AppSettingsModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<AppSettings>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<AppSettings>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => AppSettingsModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<AppSettings>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<AppSettings>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => AppSettingsModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
