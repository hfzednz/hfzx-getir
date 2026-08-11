import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/about_entity.dart';
import '../../domain/repositories/about_repository.dart';
import '../datasources/about_remote_datasource.dart';

class AboutRepositoryImpl implements AboutRepository {
  const AboutRepositoryImpl(this._remote);
  final AboutRemoteDataSource _remote;

  @override
  Future<Result<AboutInfo>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<AboutInfo>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<AboutInfo>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
