import 'package:equatable/equatable.dart';

class PerformanceMetrics extends Equatable {
  const PerformanceMetrics({
    this.acceptanceRate = 0,
    this.completionRate = 0,
    this.onTimeRate = 0,
    this.rating = 0,
    this.safetyScore = 0,
  });

  final double acceptanceRate;
  final double completionRate;
  final double onTimeRate;
  final double rating;
  final double safetyScore;

  factory PerformanceMetrics.fromJson(Map<String, dynamic> json) {
    return PerformanceMetrics(
      acceptanceRate: (json['acceptance_rate'] as num?)?.toDouble() ?? 0,
      completionRate: (json['completion_rate'] as num?)?.toDouble() ?? 0,
      onTimeRate: (json['on_time_rate'] as num?)?.toDouble() ?? 0,
      rating: (json['rating'] as num?)?.toDouble() ?? 0,
      safetyScore: (json['safety_score'] as num?)?.toDouble() ?? 0,
    );
  }

  @override
  List<Object?> get props =>
      [acceptanceRate, completionRate, onTimeRate, rating, safetyScore];
}
