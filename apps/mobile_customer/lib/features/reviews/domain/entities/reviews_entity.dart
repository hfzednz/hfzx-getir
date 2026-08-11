import 'package:equatable/equatable.dart';

enum ReviewTargetType { product, courier, order }

class Review extends Equatable {
  const Review({
    required this.id,
    required this.targetType,
    required this.targetId,
    this.orderId,
    this.rating = 0,
    this.comment = '',
    this.photoUrls = const [],
    this.verifiedPurchase = false,
    this.createdAt,
    this.authorLabel = '',
  });

  final String id;
  final ReviewTargetType targetType;
  final String targetId;
  final String? orderId;
  final int rating;
  final String comment;
  final List<String> photoUrls;
  final bool verifiedPurchase;
  final DateTime? createdAt;
  final String authorLabel;

  factory Review.fromJson(Map<String, dynamic> json) => Review(
        id: json['id']?.toString() ?? '',
        targetType: ReviewTargetType.values.asNameMap()[json['target_type']?.toString()] ??
            ReviewTargetType.order,
        targetId: json['target_id']?.toString() ?? json['product_id']?.toString() ?? '',
        orderId: json['order_id']?.toString(),
        rating: (json['rating'] as num?)?.toInt() ?? 0,
        comment: json['comment']?.toString() ?? json['body']?.toString() ?? '',
        photoUrls: (json['photo_urls'] as List<dynamic>?)?.map((e) => e.toString()).toList() ?? const [],
        verifiedPurchase: json['verified_purchase'] as bool? ?? false,
        createdAt: json['created_at'] != null ? DateTime.tryParse(json['created_at'].toString()) : null,
        authorLabel: json['author_label']?.toString() ?? '',
      );

  @override
  List<Object?> get props => [id, targetType, targetId, rating, verifiedPurchase];
}

class ReviewSubmission extends Equatable {
  const ReviewSubmission({
    required this.orderId,
    this.productRatings = const {},
    this.courierRating,
    this.orderRating,
    this.comment = '',
    this.photoPaths = const [],
  });

  final String orderId;
  final Map<String, int> productRatings;
  final int? courierRating;
  final int? orderRating;
  final String comment;
  final List<String> photoPaths;

  Map<String, dynamic> toJson() => {
        'order_id': orderId,
        if (productRatings.isNotEmpty) 'product_ratings': productRatings,
        if (courierRating != null) 'courier_rating': courierRating,
        if (orderRating != null) 'order_rating': orderRating,
        if (comment.isNotEmpty) 'comment': comment,
        if (photoPaths.isNotEmpty) 'photo_paths': photoPaths,
      };

  @override
  List<Object?> get props => [orderId, productRatings, courierRating, orderRating];
}

class ReviewAbuseReport extends Equatable {
  const ReviewAbuseReport({
    required this.reviewId,
    required this.reason,
    this.details = '',
  });

  final String reviewId;
  final String reason;
  final String details;

  Map<String, dynamic> toJson() => {
        'review_id': reviewId,
        'reason': reason,
        if (details.isNotEmpty) 'details': details,
      };

  @override
  List<Object?> get props => [reviewId, reason];
}
