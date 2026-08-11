import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/city_entity.dart';
import '../models/city_model.dart';

class CityRemoteDataSource {
  const CityRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/cities';
  static const _mutatePath = '/city';

  Future<Result<CityContext>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<CityContext>(
      path,
      parser: (json) => CityContextModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<CityContext>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<CityContext>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => CityContextModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<CityContext>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<CityContext>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => CityContextModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
