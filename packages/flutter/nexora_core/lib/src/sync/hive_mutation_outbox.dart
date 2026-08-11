import 'package:hive_ce/hive.dart';

import 'models/pending_mutation.dart';
import 'mutation_outbox.dart';

/// In-memory index backed by Hive CE for durable mutation queue storage.
class HiveMutationOutbox implements MutationOutbox {
  HiveMutationOutbox._(this._box, this._cache);

  static const boxName = 'nexora_mutation_outbox';

  final Box<dynamic> _box;
  final Map<String, PendingMutation> _cache;

  static Future<HiveMutationOutbox> open({
    required String path,
    String boxName = HiveMutationOutbox.boxName,
  }) async {
    Hive.init(path);
    final box = await Hive.openBox<dynamic>(boxName);
    final cache = <String, PendingMutation>{};

    for (final key in box.keys) {
      final raw = box.get(key);
      if (raw is Map) {
        final mutation =
            PendingMutation.fromJson(Map<String, dynamic>.from(raw));
        cache[mutation.clientMutationId] = mutation;
      }
    }

    return HiveMutationOutbox._(box, cache);
  }

  @override
  Future<void> enqueue(PendingMutation mutation) async {
    _cache[mutation.clientMutationId] = mutation;
    await _box.put(mutation.clientMutationId, mutation.toJson());
  }

  @override
  Future<List<PendingMutation>> peek({int limit = 50}) async {
    final now = DateTime.now().toUtc();
    final ready = _cache.values
        .where(
          (mutation) =>
              mutation.nextAttemptAt == null ||
              !mutation.nextAttemptAt!.isAfter(now),
        )
        .toList()
      ..sort((a, b) => a.createdAt.compareTo(b.createdAt));

    if (ready.length <= limit) {
      return ready;
    }
    return ready.sublist(0, limit);
  }

  @override
  Future<void> markSucceeded(String clientMutationId) async {
    _cache.remove(clientMutationId);
    await _box.delete(clientMutationId);
  }

  @override
  Future<void> markFailed(
    String clientMutationId, {
    required String error,
    required DateTime nextAttemptAt,
    required int retryCount,
  }) async {
    final existing = _cache[clientMutationId];
    if (existing == null) {
      return;
    }

    final updated = existing.copyWith(
      retryCount: retryCount,
      nextAttemptAt: nextAttemptAt,
      lastError: error,
    );
    _cache[clientMutationId] = updated;
    await _box.put(clientMutationId, updated.toJson());
  }

  @override
  Future<int> length() async => _cache.length;

  @override
  Future<void> clear() async {
    _cache.clear();
    await _box.clear();
  }

  Future<void> close() => _box.close();
}

/// Ephemeral outbox for unit tests (no disk persistence).
class InMemoryMutationOutbox implements MutationOutbox {
  final Map<String, PendingMutation> _cache = {};

  @override
  Future<void> clear() async => _cache.clear();

  @override
  Future<void> enqueue(PendingMutation mutation) async {
    _cache[mutation.clientMutationId] = mutation;
  }

  @override
  Future<int> length() async => _cache.length;

  @override
  Future<void> markFailed(
    String clientMutationId, {
    required String error,
    required DateTime nextAttemptAt,
    required int retryCount,
  }) async {
    final existing = _cache[clientMutationId];
    if (existing == null) {
      return;
    }
    _cache[clientMutationId] = existing.copyWith(
      retryCount: retryCount,
      nextAttemptAt: nextAttemptAt,
      lastError: error,
    );
  }

  @override
  Future<void> markSucceeded(String clientMutationId) async {
    _cache.remove(clientMutationId);
  }

  @override
  Future<List<PendingMutation>> peek({int limit = 50}) async {
    final now = DateTime.now().toUtc();
    final ready = _cache.values
        .where(
          (mutation) =>
              mutation.nextAttemptAt == null ||
              !mutation.nextAttemptAt!.isAfter(now),
        )
        .toList()
      ..sort((a, b) => a.createdAt.compareTo(b.createdAt));

    if (ready.length <= limit) {
      return ready;
    }
    return ready.sublist(0, limit);
  }
}
