import 'package:equatable/equatable.dart';

class WarehouseShift extends Equatable {
  const WarehouseShift({
    required this.id,
    required this.status,
    this.clockedInAt,
    this.clockedOutAt,
    this.onBreak = false,
  });
  final String id;
  final String status;
  final DateTime? clockedInAt;
  final DateTime? clockedOutAt;
  final bool onBreak;
  factory WarehouseShift.fromJson(Map<String, dynamic> json) => WarehouseShift(
        id: json['id']?.toString() ?? '',
        status: json['status']?.toString() ?? 'scheduled',
        clockedInAt: DateTime.tryParse(json['clocked_in_at']?.toString() ?? ''),
        clockedOutAt: DateTime.tryParse(json['clocked_out_at']?.toString() ?? ''),
        onBreak: json['on_break'] as bool? ?? false,
      );
  @override
  List<Object?> get props => [id, status, clockedInAt, clockedOutAt, onBreak];
}
