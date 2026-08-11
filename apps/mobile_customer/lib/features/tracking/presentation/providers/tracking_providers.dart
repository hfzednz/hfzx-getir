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

/// Emits the latest tracking snapshot and refreshes when realtime events
/// mention [orderId].
final trackingRealtimeProvider =
    StreamProvider.autoDispose.family<TrackingSnapshot, String>((ref, orderId) async* {
  yield await ref.watch(trackingSnapshotProvider(orderId).future);

  final client = ref.watch(realtimeClientProvider);
  unawaited(client.connect());

  await for (final event in client.events) {
    if (event is! RealtimeMessageEvent) continue;
    final payload = event.payload;
    if (!payload.contains(orderId)) continue;
    ref.invalidate(trackingSnapshotProvider(orderId));
    yield await ref.read(trackingSnapshotProvider(orderId).future);
  }
});
