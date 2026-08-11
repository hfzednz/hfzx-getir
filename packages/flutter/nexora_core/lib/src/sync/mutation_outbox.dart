import 'models/pending_mutation.dart';

/// Durable mutation outbox for offline-first sync (CONSTITUTION §25).
abstract class MutationOutbox {
  Future<void> enqueue(PendingMutation mutation);

  Future<List<PendingMutation>> peek({int limit = 50});

  Future<void> markSucceeded(String clientMutationId);

  Future<void> markFailed(
    String clientMutationId, {
    required String error,
    required DateTime nextAttemptAt,
    required int retryCount,
  });

  Future<int> length();

  Future<void> clear();
}
