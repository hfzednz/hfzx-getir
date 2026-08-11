import '../../domain/entities/reviews_entity.dart';

class ReviewModel {
  const ReviewModel({required this.raw});
  final Map<String, dynamic> raw;
  factory ReviewModel.fromJson(Map<String, dynamic> json) => ReviewModel(raw: json);
  Review toEntity() => Review.fromJson(raw);
}
