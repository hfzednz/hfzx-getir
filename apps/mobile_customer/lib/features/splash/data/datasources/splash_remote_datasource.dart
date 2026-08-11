import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/splash_entity.dart';
import '../models/splash_model.dart';

class SplashRemoteDataSource {
  const SplashRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/splash';
  static const _mutatePath = '/splash';

  Future<Result<SplashState>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<SplashState>(
      path,
      parser: (json) => SplashStateModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<SplashState>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<SplashState>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => SplashStateModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<SplashState>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<SplashState>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => SplashStateModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
