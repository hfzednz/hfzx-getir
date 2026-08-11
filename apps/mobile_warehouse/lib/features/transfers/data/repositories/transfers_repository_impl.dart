import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/transfer_entity.dart';
import '../../domain/repositories/transfers_repository.dart';
import '../datasources/transfers_remote_datasource.dart';

class TransfersRepositoryImpl implements TransfersRepository {
  TransfersRepositoryImpl(this._remote);
  final TransfersRemoteDataSource _remote;
  @override
  Future<Result<List<WarehouseTransfer>>> list() => _remote.list();
  @override
  Future<Result<WarehouseTransfer>> create(Map<String, dynamic> payload, {required String idempotencyKey}) =>
      _remote.create(payload, idempotencyKey: idempotencyKey);
  @override
  Future<Result<WarehouseTransfer>> approve(String id, {required String idempotencyKey}) =>
      _remote.approve(id, idempotencyKey: idempotencyKey);
}
