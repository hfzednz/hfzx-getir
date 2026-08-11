import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/expiry_remote_datasource.dart';
import '../../data/repositories/expiry_repository_impl.dart';
import '../../domain/entities/expiry_item.dart';
import '../../domain/repositories/expiry_repository.dart';

final expiryRemoteDataSourceProvider = Provider((ref) => ExpiryRemoteDataSource(ref.watch(apiClientProvider)));
final expiryRepositoryProvider = Provider<ExpiryRepository>((ref) => ExpiryRepositoryImpl(ref.watch(expiryRemoteDataSourceProvider)));
final nearExpiryProvider = FutureProvider.autoDispose<List<ExpiryItem>>((ref) async {
  final r = await ref.watch(expiryRepositoryProvider).listNearExpiry();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final expiryActionsProvider = Provider((ref) => ExpiryActions(ref));

class ExpiryActions {
  ExpiryActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  Future<Result<void>> wasteRemove({required String sku, required int qty}) async {
    final r = await _ref.read(expiryRepositoryProvider).wasteRemove(sku: sku, qty: qty, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(nearExpiryProvider);
    return r;
  }
}
