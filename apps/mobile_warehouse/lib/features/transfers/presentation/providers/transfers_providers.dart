import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/transfers_remote_datasource.dart';
import '../../data/repositories/transfers_repository_impl.dart';
import '../../domain/entities/transfer_entity.dart';
import '../../domain/repositories/transfers_repository.dart';

final transfersRemoteDataSourceProvider = Provider((ref) => TransfersRemoteDataSource(ref.watch(apiClientProvider)));
final transfersRepositoryProvider = Provider<TransfersRepository>((ref) => TransfersRepositoryImpl(ref.watch(transfersRemoteDataSourceProvider)));
final transfersListProvider = FutureProvider.autoDispose<List<WarehouseTransfer>>((ref) async {
  final r = await ref.watch(transfersRepositoryProvider).list();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final transfersActionsProvider = Provider((ref) => TransfersActions(ref));

class TransfersActions {
  TransfersActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  Future<Result<WarehouseTransfer>> create(Map<String, dynamic> payload) async {
    final r = await _ref.read(transfersRepositoryProvider).create(payload, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(transfersListProvider);
    return r;
  }

  Future<Result<WarehouseTransfer>> approve(String id) async {
    final r = await _ref.read(transfersRepositoryProvider).approve(id, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(transfersListProvider);
    return r;
  }
}
