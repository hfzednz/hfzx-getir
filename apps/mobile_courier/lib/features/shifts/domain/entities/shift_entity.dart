import 'package:equatable/equatable.dart';

class CourierShift extends Equatable {
  const CourierShift({
    required this.id,
    this.startedAt,
    this.endedAt,
    this.onBreak = false,
    this.breakStartedAt,
    this.plannedMinutes = 0,
    this.workedMinutes = 0,
    this.overtime = false,
  });

  final String id;
  final DateTime? startedAt;
  final DateTime? endedAt;
  final bool onBreak;
  final DateTime? breakStartedAt;
  final int plannedMinutes;
  final int workedMinutes;
  final bool overtime;

  bool get isActive => startedAt != null && endedAt == null;

  factory CourierShift.fromJson(Map<String, dynamic> json) {
    return CourierShift(
      id: json['id']?.toString() ?? '',
      startedAt: DateTime.tryParse(json['started_at']?.toString() ?? ''),
      endedAt: DateTime.tryParse(json['ended_at']?.toString() ?? ''),
      onBreak: json['on_break'] == true,
      breakStartedAt:
          DateTime.tryParse(json['break_started_at']?.toString() ?? ''),
      plannedMinutes: (json['planned_minutes'] as num?)?.toInt() ?? 0,
      workedMinutes: (json['worked_minutes'] as num?)?.toInt() ?? 0,
      overtime: json['overtime'] == true || json['is_overtime'] == true,
    );
  }

  @override
  List<Object?> get props => [
        id,
        startedAt,
        endedAt,
        onBreak,
        breakStartedAt,
        plannedMinutes,
        workedMinutes,
        overtime,
      ];
}
