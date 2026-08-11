import 'package:equatable/equatable.dart';

class ShiftSummary extends Equatable {
  const ShiftSummary({
    this.id,
    this.startsAt,
    this.endsAt,
    this.label,
  });

  final String? id;
  final DateTime? startsAt;
  final DateTime? endsAt;
  final String? label;

  factory ShiftSummary.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const ShiftSummary();
    return ShiftSummary(
      id: json['id']?.toString(),
      startsAt: DateTime.tryParse(json['starts_at']?.toString() ?? ''),
      endsAt: DateTime.tryParse(json['ends_at']?.toString() ?? ''),
      label: json['label']?.toString(),
    );
  }

  @override
  List<Object?> get props => [id, startsAt, endsAt, label];
}

class CourierDashboard extends Equatable {
  const CourierDashboard({
    required this.todayEarningsMinor,
    this.currency = 'TRY',
    this.completedCount = 0,
    this.pendingCount = 0,
    this.acceptanceRate = 0,
    this.performanceScore = 0,
    this.aiTip,
    this.currentShift,
    this.nextShift,
    this.connectionQuality,
  });

  final int todayEarningsMinor;
  final String currency;
  final int completedCount;
  final int pendingCount;
  final double acceptanceRate;
  final double performanceScore;
  final String? aiTip;
  final ShiftSummary? currentShift;
  final ShiftSummary? nextShift;
  final String? connectionQuality;

  factory CourierDashboard.fromJson(Map<String, dynamic> json) {
    return CourierDashboard(
      todayEarningsMinor: (json['today_earnings_minor'] as num?)?.toInt() ??
          (json['earnings_today_minor'] as num?)?.toInt() ??
          0,
      currency: json['currency']?.toString() ?? 'TRY',
      completedCount: (json['completed_count'] as num?)?.toInt() ?? 0,
      pendingCount: (json['pending_count'] as num?)?.toInt() ?? 0,
      acceptanceRate: (json['acceptance_rate'] as num?)?.toDouble() ?? 0,
      performanceScore: (json['performance_score'] as num?)?.toDouble() ?? 0,
      aiTip: json['ai_tip']?.toString() ?? json['tip']?.toString(),
      currentShift: ShiftSummary.fromJson(
        json['current_shift'] as Map<String, dynamic>?,
      ),
      nextShift: ShiftSummary.fromJson(
        json['next_shift'] as Map<String, dynamic>?,
      ),
      connectionQuality: json['connection_quality']?.toString(),
    );
  }

  @override
  List<Object?> get props => [
        todayEarningsMinor,
        currency,
        completedCount,
        pendingCount,
        acceptanceRate,
        performanceScore,
        aiTip,
        currentShift,
        nextShift,
        connectionQuality,
      ];
}
