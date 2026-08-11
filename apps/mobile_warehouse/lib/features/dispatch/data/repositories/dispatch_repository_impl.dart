import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/handoff_task.dart';
import '../../domain/repositories/dispatch_repository.dart';
import '../datasources/dispatch_remote_datasource.dart';

class DispatchRepositoryImpl implements DispatchRepository {
  DispatchRepositoryImpl(this._remote);
  final DispatchRemoteDataSource _remote;

  @override
  Future<Result<List<HandoffTask>>> listQueue() => _remote.listQueue();
  @override
  Future<Result<HandoffTask>> getHandoff(String id) => _remote.getHandoff(id);
  @override
  Future<Result<HandoffTask>> markCourierArrived(String id, {required String idempotencyKey}) =>
      _remote.markCourierArrived(id, idempotencyKey: idempotencyKey);
  @override
  Future<Result<HandoffTask>> scanHandoff({
    required String id,
    required String scannedToken,
    String? scannedOrderId,
    required String idempotencyKey,
  }) =>
      _remote.scanHandoff(
        id: id,
        scannedToken: scannedToken,
        scannedOrderId: scannedOrderId,
        idempotencyKey: idempotencyKey,
      );
  @override
  Future<Result<HandoffTask>> confirm(String id, {required String idempotencyKey}) =>
      _remote.confirm(id, idempotencyKey: idempotencyKey);
  @override
  Future<Result<HandoffTask>> fail(String id, {required String reasonCode, String? notes, required String idempotencyKey}) =>
      _remote.fail(id, reasonCode: reasonCode, notes: notes, idempotencyKey: idempotencyKey);
}
