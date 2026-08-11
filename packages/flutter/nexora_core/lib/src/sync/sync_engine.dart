import '../errors/nexora_exception.dart';
import 'models/pending_mutation.dart';

/// Executes pending mutations against the server when connectivity returns.
abstract class SyncEngine {
  Future<void> flush();

  Stream<SyncEngineEvent> get events;
}

sealed class SyncEngineEvent {
  const SyncEngineEvent();
}

final class SyncMutationSucceeded extends SyncEngineEvent {
  const SyncMutationSucceeded(this.mutation);

  final PendingMutation mutation;
}

final class SyncMutationFailed extends SyncEngineEvent {
  const SyncMutationFailed(this.mutation, this.error);

  final PendingMutation mutation;
  final NexoraException error;
}

final class SyncFlushCompleted extends SyncEngineEvent {
  const SyncFlushCompleted({required this.processed, required this.remaining});

  final int processed;
  final int remaining;
}
