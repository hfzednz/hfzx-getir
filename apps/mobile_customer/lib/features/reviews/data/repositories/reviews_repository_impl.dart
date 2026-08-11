import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/reviews_entity.dart';
import '../../domain/repositories/reviews_repository.dart';
import '../datasources/reviews_remote_datasource.dart';

class ReviewsRepositoryImpl implements ReviewsRepository {
  const ReviewsRepositoryImpl(this._remote);
  final ReviewsRemoteDataSource _remote;

  @override
  Future<Result<List<Review>>> fetchForOrder(String orderId) => _remote.fetchForOrder(orderId);

  @override
  Future<Result<Review>> submitReview({
    required ReviewSubmission submission,
    List<String>? photoUrls,
    String? idempotencyKey,
  }) =>
      _remote.submitReview(submission: submission, photoUrls: photoUrls, idempotencyKey: idempotencyKey);

  @override
  Future<Result<void>> reportAbuse(ReviewAbuseReport report) => _remote.reportAbuse(report);
}
