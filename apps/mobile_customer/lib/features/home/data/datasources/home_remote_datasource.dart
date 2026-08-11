import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/home_entity.dart';
import '../models/home_model.dart';

class HomeRemoteDataSource {
  const HomeRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/home';
  static const _mutatePath = '/home';

  Future<Result<HomeFeed>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<HomeFeed>(
      path,
      parser: (json) => HomeFeedModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<HomeFeed>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<HomeFeed>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => HomeFeedModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<HomeFeed>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<HomeFeed>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => HomeFeedModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
