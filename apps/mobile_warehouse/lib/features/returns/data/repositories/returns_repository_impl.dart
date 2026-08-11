import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/return_task.dart';
import '../../domain/repositories/returns_repository.dart';
import '../datasources/returns_remote_datasource.dart';

class ReturnsRepositoryImpl implements ReturnsRepository {
  ReturnsRepositoryImpl(this._remote);
  final ReturnsRemoteDataSource _remote;
  @override
  Future<Result<List<ReturnTask>>> list({String? type}) => _remote.list(type: type);
  @override
  Future<Result<ReturnTask>> advance(String id, {required String action, required String idempotencyKey}) =>
      _remote.advance(id, action: action, idempotencyKey: idempotencyKey);
}
