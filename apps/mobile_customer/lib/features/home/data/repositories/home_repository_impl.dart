import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/home_entity.dart';
import '../../domain/repositories/home_repository.dart';
import '../datasources/home_remote_datasource.dart';

class HomeRepositoryImpl implements HomeRepository {
  const HomeRepositoryImpl(this._remote);
  final HomeRemoteDataSource _remote;

  @override
  Future<Result<HomeFeed>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<HomeFeed>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<HomeFeed>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
