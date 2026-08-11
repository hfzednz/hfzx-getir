import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/legal_entity.dart';
import '../../domain/repositories/legal_repository.dart';
import '../datasources/legal_remote_datasource.dart';

class LegalRepositoryImpl implements LegalRepository {
  const LegalRepositoryImpl(this._remote);
  final LegalRemoteDataSource _remote;

  @override
  Future<Result<LegalDocument>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<LegalDocument>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<LegalDocument>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
