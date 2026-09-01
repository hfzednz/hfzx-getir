import 'dart:async';
import 'dart:convert';

import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../../shared/realtime/order_sse_client.dart';
import '../../../../shared/realtime/sse_parser.dart';
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

/// Live tracking: authenticated HTTP SSE is primary; polling is fallback only.
final trackingRealtimeProvider =
    StreamProvider.autoDispose.family<TrackingSnapshot, String>((ref, orderId) {
  final controller = StreamController<TrackingSnapshot>();
  var closed = false;
  var sseConnected = false;
  var lastRank = -1;
  final seenEventIds = <String>{};
  TrackingSnapshot? latest;

  final repo = ref.read(trackingRepositoryProvider);
  final env = ref.read(environmentProvider);
  final sse = OrderSseClient();
  StreamSubscription<SseFrame>? sseSub;
  Timer? pollTimer;
  AppLifecycleListener? lifecycle;

  void emit(TrackingSnapshot next) {
    final rank = trackingStatusRank(next.status);
    if (rank >= 0 && lastRank >= 0 && rank < lastRank) return;
    if (rank >= 0) lastRank = rank;
    latest = next;
    if (!controller.isClosed) controller.add(next);
  }

  Future<void> refreshTrack() async {
    ref.invalidate(trackingSnapshotProvider(orderId));
    try {
      emit(await ref.read(trackingSnapshotProvider(orderId).future));
    } catch (_) {}
  }

  void handleFrame(SseFrame frame) {
    sseConnected = true;
    if (frame.id != null && frame.id!.isNotEmpty) {
      if (!seenEventIds.add(frame.id!)) return;
      if (seenEventIds.length > 200) {
        seenEventIds.remove(seenEventIds.first);
      }
    }
    if (frame.data.isEmpty) return;
    try {
      final decoded = jsonDecode(frame.data);
      if (decoded is Map && latest != null) {
        final status = decoded['status']?.toString() ??
            decoded['type']?.toString() ??
            decoded['event']?.toString();
        if (status != null && status.isNotEmpty) {
          final current = latest!;
          emit(
            TrackingSnapshot(
              orderId: current.orderId,
              status: normalizeTrackingStatus(status),
              etaMinutes: current.etaMinutes,
              etaMin: current.etaMin,
              etaMax: current.etaMax,
              courierName: current.courierName,
              courierPhone: current.courierPhone,
              courierLat: current.courierLat,
              courierLng: current.courierLng,
              storeLat: current.storeLat,
              storeLng: current.storeLng,
              destLat: current.destLat,
              destLng: current.destLng,
              steps: trackingLifecycleSteps(status),
              routePoints: current.routePoints,
              canCall: current.canCall,
              canChat: current.canChat,
              courierChatUrl: current.courierChatUrl,
            ),
          );
        }
      }
    } catch (_) {}
    unawaited(refreshTrack());
  }

  Future<void> connectSse() async {
    if (closed) return;
    await sseSub?.cancel();
    final ticketResult = await repo.issueRealtimeTicket(orderId);
    final ticket = ticketResult.fold(
      onSuccess: (value) => value,
      onFailure: (_) => const RealtimeTicket(ticket: ''),
    );
    if (ticket.isEmpty || closed) return;
    final topic = ticket.topic.isEmpty ? 'order:$orderId' : ticket.topic;
    sseSub = sse
        .connect(sseUrl: env.sseUrl, ticket: ticket.ticket, topic: topic)
        .listen(
      handleFrame,
      onError: (_) {
        sseConnected = false;
      },
    );
  }

  void startPollFallback() {
    pollTimer?.cancel();
    pollTimer = Timer.periodic(const Duration(seconds: 4), (_) {
      if (closed) return;
      if (sseConnected) return;
      unawaited(connectSse());
      unawaited(refreshTrack());
    });
  }

  unawaited(() async {
    try {
      emit(await ref.read(trackingSnapshotProvider(orderId).future));
    } catch (err, stack) {
      if (!controller.isClosed) controller.addError(err, stack);
    }
    await connectSse();
    startPollFallback();
  }());

  lifecycle = AppLifecycleListener(
    onResume: () {
      sseConnected = false;
      unawaited(connectSse());
      unawaited(refreshTrack());
    },
    onPause: () {
      unawaited(sseSub?.cancel());
      sseConnected = false;
    },
  );

  ref.onDispose(() {
    closed = true;
    lifecycle?.dispose();
    pollTimer?.cancel();
    unawaited(sseSub?.cancel());
    unawaited(controller.close());
  });

  return controller.stream;
});
