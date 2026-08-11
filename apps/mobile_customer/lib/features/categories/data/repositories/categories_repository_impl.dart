import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/categories_entity.dart';
import '../../domain/repositories/categories_repository.dart';
import '../datasources/categories_remote_datasource.dart';

class CategoriesRepositoryImpl implements CategoriesRepository {
  const CategoriesRepositoryImpl(this._remote);
  final CategoriesRemoteDataSource _remote;

  @override
  Future<Result<Category>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<Category>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Category>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
