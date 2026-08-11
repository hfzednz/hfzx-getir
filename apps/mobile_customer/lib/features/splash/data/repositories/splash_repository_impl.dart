import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/splash_entity.dart';
import '../../domain/repositories/splash_repository.dart';
import '../datasources/splash_remote_datasource.dart';

class SplashRepositoryImpl implements SplashRepository {
  const SplashRepositoryImpl(this._remote);
  final SplashRemoteDataSource _remote;

  @override
  Future<Result<SplashState>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<SplashState>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<SplashState>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
