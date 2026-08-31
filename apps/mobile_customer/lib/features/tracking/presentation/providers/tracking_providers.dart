import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/tracking_remote_datasource.dart';
import '../../data/repositories/tracking_repository_impl.dart';
import '../../domain/entities/tracking_entity.dart';
import '../../domain/repositories/tracking_repository.dart';
import '../../domain/usecases/tracking_usecases.dart';

final trackingRemoteDataSourceProvider = Provider<TrackingRemoteDataSource>((ref) {
  return TrackingRemoteDataSource(ref.watch(apiClientProvider));
});

final trackingRepositoryProvider = Provider<TrackingRepository>((ref) {
  return TrackingRepositoryImpl(ref.watch(trackingRemoteDataSourceProvider));
});

final getTrackingSnapshotUseCaseProvider = Provider(
  (ref) => GetTrackingSnapshotUseCase(ref.watch(trackingRepositoryProvider)),
);

final trackingSnapshotProvider =
    FutureProvider.autoDispose.family<TrackingSnapshot, String>((ref, orderId) async {
  final result = await ref.watch(getTrackingSnapshotUseCaseProvider).call(orderId);
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});

/// Live tracking: authenticated SSE/WS ticket for this order, plus track polling.
/// Warehouse and courier events arrive on the ticketed realtime socket; poll is
/// the fallback when the socket is reconnecting.
final trackingRealtimeProvider =
    StreamProvider.autoDispose.family<TrackingSnapshot, String>((ref, orderId) async* {
  yield await ref.watch(trackingSnapshotProvider(orderId).future);

  var closed = false;
  ref.onDispose(() => closed = true);

  final repo = ref.watch(trackingRepositoryProvider);
  final env = ref.watch(environmentProvider);
  final ticketResult = await repo.issueRealtimeTicket(orderId);
  final ticket = ticketResult.fold(
    onSuccess: (value) => value,
    onFailure: (_) => '',
  );

  RealtimeClient? live;
  StreamSubscription<RealtimeEvent>? sub;
  if (ticket.isNotEmpty) {
    live = RealtimeClient(
      wsBaseUrl: env.wsUrl,
      ticketProvider: () async => ticket,
    );
    sub = live.events.listen((event) {
      if (event is RealtimeMessageEvent && !closed) {
        ref.invalidate(trackingSnapshotProvider(orderId));
      }
    });
    ref.onDispose(() {
      unawaited(sub?.cancel());
      unawaited(live?.dispose());
    });
    unawaited(live.connect());
  }

  await for (final _ in Stream<void>.periodic(const Duration(seconds: 3))) {
    if (closed) return;
    ref.invalidate(trackingSnapshotProvider(orderId));
    try {
      yield await ref.read(trackingSnapshotProvider(orderId).future);
    } catch (_) {
      // Keep the last snapshot on transient track failures.
    }
  }
});
