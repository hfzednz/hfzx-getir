import 'package:nexora_core/nexora_core.dart';

import '../entities/reviews_entity.dart';
import '../repositories/reviews_repository.dart';

class ListOrderReviewsUseCase {
  const ListOrderReviewsUseCase(this._repository);
  final ReviewsRepository _repository;

  Future<Result<List<Review>>> call(String orderId) => _repository.fetchForOrder(orderId);
}

class SubmitReviewUseCase {
  const SubmitReviewUseCase(this._repository);
  final ReviewsRepository _repository;

  Future<Result<Review>> call({
    required ReviewSubmission submission,
    List<String>? photoUrls,
    String? idempotencyKey,
  }) =>
      _repository.submitReview(
        submission: submission,
        photoUrls: photoUrls,
        idempotencyKey: idempotencyKey,
      );
}
