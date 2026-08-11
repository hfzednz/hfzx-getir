import '../../domain/entities/tracking_entity.dart';

class TrackingSnapshotModel {
  const TrackingSnapshotModel({required this.entity});

  final TrackingSnapshot entity;

  factory TrackingSnapshotModel.fromJson(Map<String, dynamic> json) =>
      TrackingSnapshotModel(entity: TrackingSnapshot.fromJson(json));

  TrackingSnapshot toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}
