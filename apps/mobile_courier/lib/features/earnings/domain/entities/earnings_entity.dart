import 'package:equatable/equatable.dart';

enum EarningsPeriod { daily, weekly, monthly }

class EarningsBreakdown extends Equatable {
  const EarningsBreakdown({
    this.baseMinor = 0,
    this.tipsMinor = 0,
    this.bonusesMinor = 0,
    this.penaltiesMinor = 0,
    this.currency = 'TRY',
  });

  final int baseMinor;
  final int tipsMinor;
  final int bonusesMinor;
  final int penaltiesMinor;
  final String currency;

  int get netMinor => baseMinor + tipsMinor + bonusesMinor - penaltiesMinor;

  factory EarningsBreakdown.fromJson(Map<String, dynamic> json) {
    return EarningsBreakdown(
      baseMinor: (json['base_minor'] as num?)?.toInt() ?? 0,
      tipsMinor: (json['tips_minor'] as num?)?.toInt() ?? 0,
      bonusesMinor: (json['bonuses_minor'] as num?)?.toInt() ?? 0,
      penaltiesMinor: (json['penalties_minor'] as num?)?.toInt() ?? 0,
      currency: json['currency']?.toString() ?? 'TRY',
    );
  }

  @override
  List<Object?> get props =>
      [baseMinor, tipsMinor, bonusesMinor, penaltiesMinor, currency];
}

class PayoutRecord extends Equatable {
  const PayoutRecord({
    required this.id,
    required this.amountMinor,
    required this.currency,
    required this.status,
    this.paidAt,
  });

  final String id;
  final int amountMinor;
  final String currency;
  final String status;
  final DateTime? paidAt;

  factory PayoutRecord.fromJson(Map<String, dynamic> json) {
    return PayoutRecord(
      id: json['id']?.toString() ?? '',
      amountMinor: (json['amount_minor'] as num?)?.toInt() ?? 0,
      currency: json['currency']?.toString() ?? 'TRY',
      status: json['status']?.toString() ?? 'pending',
      paidAt: DateTime.tryParse(json['paid_at']?.toString() ?? ''),
    );
  }

  @override
  List<Object?> get props => [id, amountMinor, currency, status, paidAt];
}

class EarningsSnapshot extends Equatable {
  const EarningsSnapshot({
    required this.period,
    required this.breakdown,
    this.payouts = const [],
  });

  final EarningsPeriod period;
  final EarningsBreakdown breakdown;
  final List<PayoutRecord> payouts;

  factory EarningsSnapshot.fromJson(
    Map<String, dynamic> json, {
    required EarningsPeriod period,
  }) {
    final payoutsRaw = json['payouts'] as List? ?? json['payout_history'] as List? ?? [];
    return EarningsSnapshot(
      period: period,
      breakdown: EarningsBreakdown.fromJson(json),
      payouts: payoutsRaw
          .map((e) => PayoutRecord.fromJson(Map<String, dynamic>.from(e as Map)))
          .toList(),
    );
  }

  @override
  List<Object?> get props => [period, breakdown, payouts];
}
