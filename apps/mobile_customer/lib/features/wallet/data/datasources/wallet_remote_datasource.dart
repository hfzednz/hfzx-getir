import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/wallet_entity.dart';
import '../models/wallet_model.dart';

class WalletRemoteDataSource {
  const WalletRemoteDataSource(this._client);
  final ApiClient _client;

  static const _accountPath = '/wallet/account';
  static const _transactionsPath = '/wallet/transactions';
  static const _topUpPath = '/wallet/top-up';

  Future<Result<WalletAccount>> fetchAccount() async {
    return _client.get<WalletAccount>(
      _accountPath,
      parser: (json) => WalletAccountModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<WalletTransaction>>> fetchTransactions({
    Map<String, dynamic>? params,
  }) async {
    return _client.get<List<WalletTransaction>>(
      _transactionsPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => WalletTransactionModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<WalletAccount>> topUp({
    required int amountMinor,
    String currency = 'TRY',
    String? idempotencyKey,
  }) async {
    return _client.post<WalletAccount>(
      _topUpPath,
      data: {'amount_minor': amountMinor, 'currency': currency},
      idempotencyKey: idempotencyKey,
      parser: (json) => WalletAccountModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  /// Legacy list endpoint — maps to account when no list API exists.
  Future<Result<WalletAccount>> fetch({String? id}) async => fetchAccount();

  Future<Result<List<WalletAccount>>> fetchList({Map<String, dynamic>? params}) async {
    final account = await fetchAccount();
    return account.map((a) => [a]);
  }
}
