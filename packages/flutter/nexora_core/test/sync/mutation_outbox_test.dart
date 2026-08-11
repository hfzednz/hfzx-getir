import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';

void main() {
  group('InMemoryMutationOutbox', () {
    late InMemoryMutationOutbox outbox;

    setUp(() {
      outbox = InMemoryMutationOutbox();
    });

    PendingMutation mutation({
      required String id,
      DateTime? nextAttemptAt,
      int retryCount = 0,
    }) {
      return PendingMutation(
        clientMutationId: id,
        idempotencyKey: 'idem-$id',
        method: 'POST',
        path: '/v1/cart/items',
        body: {'sku': 'SKU-1', 'qty': 1},
        createdAt: DateTime.utc(2026, 8, 5, 12, 0, 0),
        nextAttemptAt: nextAttemptAt,
        retryCount: retryCount,
      );
    }

    test('enqueue and peek returns FIFO ready mutations', () async {
      await outbox.enqueue(mutation(id: 'm1'));
      await outbox.enqueue(mutation(id: 'm2'));

      final pending = await outbox.peek();

      expect(pending, hasLength(2));
      expect(pending.first.clientMutationId, 'm1');
      expect(pending.last.clientMutationId, 'm2');
    });

    test('peek excludes mutations scheduled for future retry', () async {
      await outbox.enqueue(mutation(id: 'ready'));
      await outbox.enqueue(
        mutation(
          id: 'delayed',
          nextAttemptAt: DateTime.now().toUtc().add(const Duration(hours: 1)),
        ),
      );

      final pending = await outbox.peek();

      expect(pending, hasLength(1));
      expect(pending.single.clientMutationId, 'ready');
    });

    test('markSucceeded removes mutation', () async {
      await outbox.enqueue(mutation(id: 'm1'));
      await outbox.markSucceeded('m1');

      expect(await outbox.length(), 0);
      expect(await outbox.peek(), isEmpty);
    });

    test('markFailed updates retry metadata', () async {
      await outbox.enqueue(mutation(id: 'm1'));
      final nextAttempt = DateTime.now().toUtc().add(const Duration(seconds: 30));

      await outbox.markFailed(
        'm1',
        error: 'network',
        nextAttemptAt: nextAttempt,
        retryCount: 1,
      );

      final pending = await outbox.peek();
      expect(pending, isEmpty);

      final all = await outbox.peek(limit: 10);
      expect(all, isEmpty);

      // Force-include by waiting — mutation should still exist with future nextAttemptAt
      expect(await outbox.length(), 1);
    });

    test('clear removes all mutations', () async {
      await outbox.enqueue(mutation(id: 'm1'));
      await outbox.enqueue(mutation(id: 'm2'));
      await outbox.clear();

      expect(await outbox.length(), 0);
    });
  });

  group('HiveMutationOutbox persistence', () {
    test('round-trips mutations through JSON', () {
      final original = PendingMutation(
        clientMutationId: 'mut-1',
        idempotencyKey: 'idem-1',
        method: 'PATCH',
        path: '/v1/cart/items/abc',
        body: {'qty': 2},
        createdAt: DateTime.utc(2026, 8, 5, 10, 0, 0),
        retryCount: 2,
        nextAttemptAt: DateTime.utc(2026, 8, 5, 10, 5, 0),
        lastError: '503',
      );

      final restored = PendingMutation.fromJson(original.toJson());

      expect(restored, original);
    });
  });
}
