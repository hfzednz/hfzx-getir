import 'package:nexora_core/nexora_core.dart';

import '../entities/wallet_entity.dart';
import '../repositories/wallet_repository.dart';

class GetWalletAccountUseCase {
  const GetWalletAccountUseCase(this._repository);
  final WalletRepository _repository;

  Future<Result<WalletAccount>> call() => _repository.fetchAccount();
}

class ListWalletTransactionsUseCase {
  const ListWalletTransactionsUseCase(this._repository);
  final WalletRepository _repository;

  Future<Result<List<WalletTransaction>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchTransactions(params: params);
}

class TopUpWalletUseCase {
  const TopUpWalletUseCase(this._repository);
  final WalletRepository _repository;

  Future<Result<WalletAccount>> call({
    required int amountMinor,
    String currency = 'TRY',
    String? idempotencyKey,
  }) =>
      _repository.topUp(
        amountMinor: amountMinor,
        currency: currency,
        idempotencyKey: idempotencyKey,
      );
}

/// Legacy aliases.
typedef GetWalletUseCase = GetWalletAccountUseCase;
typedef ListWalletUseCase = ListWalletTransactionsUseCase;
