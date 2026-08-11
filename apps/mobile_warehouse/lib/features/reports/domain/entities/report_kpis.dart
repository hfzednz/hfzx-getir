import 'package:equatable/equatable.dart';

class ReportKpis extends Equatable {
  const ReportKpis({
    this.pickSpeedPerHour = 0,
    this.pickAccuracy = 0,
    this.wasteUnits = 0,
    this.packSpeedPerHour = 0,
  });
  final double pickSpeedPerHour;
  final double pickAccuracy;
  final int wasteUnits;
  final double packSpeedPerHour;
  factory ReportKpis.fromJson(Map<String, dynamic> json) => ReportKpis(
        pickSpeedPerHour: (json['pick_speed_per_hour'] as num?)?.toDouble() ?? 0,
        pickAccuracy: (json['pick_accuracy'] as num?)?.toDouble() ?? 0,
        wasteUnits: (json['waste_units'] as num?)?.toInt() ?? 0,
        packSpeedPerHour: (json['pack_speed_per_hour'] as num?)?.toDouble() ?? 0,
      );
  @override
  List<Object?> get props => [pickSpeedPerHour, pickAccuracy, wasteUnits, packSpeedPerHour];
}
