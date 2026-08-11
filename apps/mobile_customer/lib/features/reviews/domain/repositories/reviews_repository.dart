import 'package:nexora_core/nexora_core.dart';

import '../entities/reviews_entity.dart';

abstract class ReviewsRepository {
  Future<Result<List<Review>>> fetchForOrder(String orderId);
  Future<Result<Review>> submitReview({
    required ReviewSubmission submission,
    List<String>? photoUrls,
    String? idempotencyKey,
  });
  Future<Result<void>> reportAbuse(ReviewAbuseReport report);
}
