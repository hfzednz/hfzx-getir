import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../data/datasources/reviews_remote_datasource.dart';
import '../../data/repositories/reviews_repository_impl.dart';
import '../../domain/entities/reviews_entity.dart';
import '../../domain/repositories/reviews_repository.dart';

final reviewsRemoteDataSourceProvider = Provider<ReviewsRemoteDataSource>((ref) {
  return ReviewsRemoteDataSource(ref.watch(apiClientProvider));
});

final reviewsRepositoryProvider = Provider<ReviewsRepository>((ref) {
  return ReviewsRepositoryImpl(ref.watch(reviewsRemoteDataSourceProvider));
});

final orderReviewsProvider =
    FutureProvider.autoDispose.family<List<Review>, String>((ref, orderId) async {
  final result = await ref.watch(reviewsRepositoryProvider).fetchForOrder(orderId);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final reviewSubmitProvider =
    AsyncNotifierProvider<ReviewSubmitController, Review?>(ReviewSubmitController.new);

class ReviewSubmitController extends AsyncNotifier<Review?> {
  @override
  Future<Review?> build() async => null;

  Future<void> submit({
    required String orderId,
    required int orderRating,
    int? courierRating,
    Map<String, int> productRatings = const {},
    required String comment,
    List<XFile>? photos,
    bool verifiedPurchase = true,
  }) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final photoPaths = photos?.map((p) => p.path).toList() ?? const [];
      final submission = ReviewSubmission(
        orderId: orderId,
        orderRating: orderRating,
        courierRating: courierRating,
        productRatings: productRatings,
        comment: comment,
        photoPaths: photoPaths,
      );

      final result = await ref.read(reviewsRepositoryProvider).submitReview(
            submission: submission,
            photoUrls: photoPaths.isEmpty ? null : photoPaths,
            idempotencyKey: const Uuid().v4(),
          );

      return result.fold(
        onSuccess: (review) {
          ref.invalidate(orderReviewsProvider(orderId));
          ref.read(analyticsTrackerProvider).trackRaw(
                eventName: AnalyticsEvents.reviewSubmitted,
                props: {
                  'order_id': orderId,
                  'rating': orderRating,
                  'verified_purchase': verifiedPurchase,
                  'photo_count': photoPaths.length,
                },
              );
          return review;
        },
        onFailure: (e) => throw e,
      );
    });
  }
}
