import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/wallet_entity.dart';
import '../../domain/repositories/wallet_repository.dart';
import '../datasources/wallet_remote_datasource.dart';

class WalletRepositoryImpl implements WalletRepository {
  const WalletRepositoryImpl(this._remote);
  final WalletRemoteDataSource _remote;

  @override
  Future<Result<WalletAccount>> fetchAccount() => _remote.fetchAccount();

  @override
  Future<Result<List<WalletTransaction>>> fetchTransactions({Map<String, dynamic>? params}) =>
      _remote.fetchTransactions(params: params);

  @override
  Future<Result<WalletAccount>> topUp({
    required int amountMinor,
    String currency = 'TRY',
    String? idempotencyKey,
  }) =>
      _remote.topUp(amountMinor: amountMinor, currency: currency, idempotencyKey: idempotencyKey);

  @override
  Future<Result<WalletAccount>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<WalletAccount>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);
}
