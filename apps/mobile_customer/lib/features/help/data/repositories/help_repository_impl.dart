import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/help_entity.dart';
import '../../domain/repositories/help_repository.dart';
import '../datasources/help_remote_datasource.dart';

class HelpRepositoryImpl implements HelpRepository {
  const HelpRepositoryImpl(this._remote);
  final HelpRemoteDataSource _remote;

  @override
  Future<Result<HelpArticle>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<HelpArticle>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<HelpArticle>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
