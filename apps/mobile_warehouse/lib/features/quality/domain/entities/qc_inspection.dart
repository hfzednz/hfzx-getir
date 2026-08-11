import 'package:equatable/equatable.dart';

class QcInspection extends Equatable {
  const QcInspection({
    required this.id,
    required this.stage,
    required this.status,
    this.reference,
  });
  final String id;
  final String stage; // receiving | packing | dispatch
  final String status;
  final String? reference;
  factory QcInspection.fromJson(Map<String, dynamic> json) => QcInspection(
        id: json['id']?.toString() ?? '',
        stage: json['stage']?.toString() ?? 'receiving',
        status: json['status']?.toString() ?? 'open',
        reference: json['reference']?.toString(),
      );
  @override
  List<Object?> get props => [id, stage, status, reference];
}
