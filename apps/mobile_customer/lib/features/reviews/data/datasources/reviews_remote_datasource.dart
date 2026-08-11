import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/reviews_entity.dart';
import '../models/reviews_model.dart';

class ReviewsRemoteDataSource {
  const ReviewsRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/reviews';
  static const _reportPath = '/reviews/report';

  Future<Result<List<Review>>> fetchForOrder(String orderId) async {
    return _client.get<List<Review>>(
      _listPath,
      queryParameters: {'order_id': orderId},
      parser: (json) => (json as List<dynamic>)
          .map((e) => ReviewModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Review>> submitReview({
    required ReviewSubmission submission,
    List<String>? photoUrls,
    String? idempotencyKey,
  }) async {
    return _client.post<Review>(
      '/orders/${submission.orderId}/review',
      data: {
        ...submission.toJson(),
        if (photoUrls != null && photoUrls.isNotEmpty) 'photo_urls': photoUrls,
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => ReviewModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<void>> reportAbuse(ReviewAbuseReport report) async {
    return _client.post<void>(
      _reportPath,
      data: report.toJson(),
      parser: (_) {},
    );
  }
}
