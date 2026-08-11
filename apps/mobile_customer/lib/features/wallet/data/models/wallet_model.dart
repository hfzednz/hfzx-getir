import '../../domain/entities/wallet_entity.dart';

class WalletAccountModel {
  const WalletAccountModel({required this.raw});

  final Map<String, dynamic> raw;

  factory WalletAccountModel.fromJson(Map<String, dynamic> json) => WalletAccountModel(raw: json);

  WalletAccount toEntity() => WalletAccount.fromJson(raw);
}

class WalletTransactionModel {
  const WalletTransactionModel({required this.raw});

  final Map<String, dynamic> raw;

  factory WalletTransactionModel.fromJson(Map<String, dynamic> json) =>
      WalletTransactionModel(raw: json);

  WalletTransaction toEntity() => WalletTransaction.fromJson(raw);
}

/// Backward-compatible alias.
typedef WalletModel = WalletAccountModel;
