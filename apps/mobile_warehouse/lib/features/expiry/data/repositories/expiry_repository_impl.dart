import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/expiry_item.dart';
import '../../domain/repositories/expiry_repository.dart';
import '../datasources/expiry_remote_datasource.dart';

class ExpiryRepositoryImpl implements ExpiryRepository {
  ExpiryRepositoryImpl(this._remote);
  final ExpiryRemoteDataSource _remote;
  @override
  Future<Result<List<ExpiryItem>>> listNearExpiry() => _remote.listNearExpiry();
  @override
  Future<Result<void>> wasteRemove({required String sku, required int qty, required String idempotencyKey}) =>
      _remote.wasteRemove(sku: sku, qty: qty, idempotencyKey: idempotencyKey);
}
