import '../../domain/entities/home_entity.dart';

class HomeFeedModel {
  const HomeFeedModel({required this.entity});

  final HomeFeed entity;

  factory HomeFeedModel.fromJson(Map<String, dynamic> json) =>
      HomeFeedModel(entity: HomeFeed.fromJson(json));

  HomeFeed toEntity() => entity;

  Map<String, dynamic> toJson() => entity.toJson();
}
