import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/city_entity.dart';
import '../../domain/repositories/city_repository.dart';
import '../datasources/city_remote_datasource.dart';

class CityRepositoryImpl implements CityRepository {
  const CityRepositoryImpl(this._remote);
  final CityRemoteDataSource _remote;

  @override
  Future<Result<CityContext>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<CityContext>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<CityContext>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
