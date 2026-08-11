import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/categories_entity.dart';
import '../models/categories_model.dart';

class CategoriesRemoteDataSource {
  const CategoriesRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/categories';
  static const _mutatePath = '/categories';

  Future<Result<Category>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<Category>(
      path,
      parser: (json) => CategoryModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<Category>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Category>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => CategoryModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Category>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Category>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => CategoryModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
