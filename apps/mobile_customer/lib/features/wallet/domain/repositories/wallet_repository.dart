import 'package:nexora_core/nexora_core.dart';

import '../entities/wallet_entity.dart';

abstract class WalletRepository {
  Future<Result<WalletAccount>> fetchAccount();
  Future<Result<List<WalletTransaction>>> fetchTransactions({Map<String, dynamic>? params});
  Future<Result<WalletAccount>> topUp({
    required int amountMinor,
    String currency,
    String? idempotencyKey,
  });

  /// Legacy compatibility.
  Future<Result<WalletAccount>> fetch({String? id});
  Future<Result<List<WalletAccount>>> fetchList({Map<String, dynamic>? params});
}
