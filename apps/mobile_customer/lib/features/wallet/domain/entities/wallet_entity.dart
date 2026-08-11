import 'package:equatable/equatable.dart';

import '../../../../shared/utils/money.dart';

enum WalletTransactionType {
  topUp,
  cashback,
  refund,
  promoCredit,
  payment,
  adjustment,
  unknown;

  static WalletTransactionType fromString(String? raw) => switch (raw?.toLowerCase()) {
        'top_up' || 'topup' => WalletTransactionType.topUp,
        'cashback' => WalletTransactionType.cashback,
        'refund' => WalletTransactionType.refund,
        'promo_credit' || 'promo' => WalletTransactionType.promoCredit,
        'payment' => WalletTransactionType.payment,
        'adjustment' => WalletTransactionType.adjustment,
        _ => WalletTransactionType.unknown,
      };

  String get label => switch (this) {
        WalletTransactionType.topUp => 'Top-up',
        WalletTransactionType.cashback => 'Cashback',
        WalletTransactionType.refund => 'Refund',
        WalletTransactionType.promoCredit => 'Promo credit',
        WalletTransactionType.payment => 'Payment',
        WalletTransactionType.adjustment => 'Adjustment',
        WalletTransactionType.unknown => 'Transaction',
      };
}

class WalletAccount extends Equatable {
  const WalletAccount({
    required this.id,
    this.balanceMinor = 0,
    this.currency = 'TRY',
    this.pendingMinor = 0,
    this.cashbackMinor = 0,
    this.promoCreditMinor = 0,
    this.topUpEnabled = true,
    this.updatedAt,
  });

  final String id;
  final int balanceMinor;
  final String currency;
  final int pendingMinor;
  final int cashbackMinor;
  final int promoCreditMinor;
  final bool topUpEnabled;
  final DateTime? updatedAt;

  Money get balance => Money(minorUnits: balanceMinor, currency: currency);

  factory WalletAccount.fromJson(Map<String, dynamic> json) {
    final balance = json['balance'] as Map<String, dynamic>?;
    return WalletAccount(
      id: json['id']?.toString() ?? json['account_id']?.toString() ?? '',
      balanceMinor: (balance?['minor_units'] as num?)?.toInt() ??
          (json['balance_minor'] as num?)?.toInt() ??
          0,
      currency: balance?['currency']?.toString() ?? json['currency']?.toString() ?? 'TRY',
      pendingMinor: (json['pending_minor'] as num?)?.toInt() ?? 0,
      cashbackMinor: (json['cashback_minor'] as num?)?.toInt() ?? 0,
      promoCreditMinor: (json['promo_credit_minor'] as num?)?.toInt() ?? 0,
      topUpEnabled: json['top_up_enabled'] as bool? ?? true,
      updatedAt: json['updated_at'] != null ? DateTime.tryParse(json['updated_at'].toString()) : null,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'balance_minor': balanceMinor,
        'currency': currency,
        'pending_minor': pendingMinor,
        'cashback_minor': cashbackMinor,
        'promo_credit_minor': promoCreditMinor,
        'top_up_enabled': topUpEnabled,
        if (updatedAt != null) 'updated_at': updatedAt!.toIso8601String(),
      };

  @override
  List<Object?> get props =>
      [id, balanceMinor, currency, pendingMinor, cashbackMinor, promoCreditMinor, topUpEnabled, updatedAt];
}

class WalletTransaction extends Equatable {
  const WalletTransaction({
    required this.id,
    required this.type,
    required this.amountMinor,
    this.currency = 'TRY',
    this.description = '',
    this.referenceId,
    this.createdAt,
    this.status = 'completed',
  });

  final String id;
  final WalletTransactionType type;
  final int amountMinor;
  final String currency;
  final String description;
  final String? referenceId;
  final DateTime? createdAt;
  final String status;

  bool get isCredit => amountMinor >= 0;

  Money get amount => Money(minorUnits: amountMinor.abs(), currency: currency);

  factory WalletTransaction.fromJson(Map<String, dynamic> json) {
    final amount = json['amount'] as Map<String, dynamic>?;
    return WalletTransaction(
      id: json['id']?.toString() ?? '',
      type: WalletTransactionType.fromString(json['type']?.toString()),
      amountMinor: (amount?['minor_units'] as num?)?.toInt() ??
          (json['amount_minor'] as num?)?.toInt() ??
          0,
      currency: amount?['currency']?.toString() ?? json['currency']?.toString() ?? 'TRY',
      description: json['description']?.toString() ?? json['title']?.toString() ?? '',
      referenceId: json['reference_id']?.toString(),
      createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
      status: json['status']?.toString() ?? 'completed',
    );
  }

  @override
  List<Object?> get props => [id, type, amountMinor, currency, description, referenceId, createdAt, status];
}

/// Backward-compatible alias used by generated scaffolding.
typedef Wallet = WalletAccount;
