import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/about_entity.dart';
import '../models/about_model.dart';

class AboutRemoteDataSource {
  const AboutRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/about';
  static const _mutatePath = '/about';

  Future<Result<AboutInfo>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<AboutInfo>(
      path,
      parser: (json) => AboutInfoModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<AboutInfo>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<AboutInfo>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => AboutInfoModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<AboutInfo>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<AboutInfo>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => AboutInfoModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
