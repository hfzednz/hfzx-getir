import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/delivery_rules.dart';
import '../../data/datasources/deliveries_remote_datasource.dart';
import '../../data/repositories/deliveries_repository_impl.dart';
import '../../domain/entities/delivery_job.dart';
import '../../domain/repositories/deliveries_repository.dart';
import '../../../status/presentation/providers/duty_controller.dart';

final deliveriesRemoteDataSourceProvider =
    Provider<DeliveriesRemoteDataSource>((ref) {
  return DeliveriesRemoteDataSource(ref.watch(apiClientProvider));
});

final deliveriesRepositoryProvider = Provider<DeliveriesRepository>((ref) {
  return DeliveriesRepositoryImpl(ref.watch(deliveriesRemoteDataSourceProvider));
});

final activeDeliveriesProvider =
    FutureProvider.autoDispose<List<DeliveryJob>>((ref) async {
  final result = await ref.watch(deliveriesRepositoryProvider).listActive();
  final jobs = result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
  ref.read(dutyControllerProvider.notifier).setHasActiveDelivery(jobs.isNotEmpty);
  return jobs;
});

final deliveryDetailProvider =
    FutureProvider.autoDispose.family<DeliveryJob, String>((ref, id) async {
  final result = await ref.watch(deliveriesRepositoryProvider).getById(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final deliveryActionsProvider = Provider((ref) => DeliveryActions(ref));

class DeliveryActions {
  DeliveryActions(this._ref);
  final Ref _ref;

  DeliveriesRepository get _repo => _ref.read(deliveriesRepositoryProvider);

  Future<DeliveryJob?> advance(DeliveryJob job, DeliveryLifecycleStatus next) async {
    final validation = DeliveryRules.validateTransition(
      from: job.status,
      to: next,
    );
    if (validation.isFailure) return null;
    final result = await _repo.transition(job.id, next);
    _ref.invalidate(activeDeliveriesProvider);
    _ref.invalidate(deliveryDetailProvider(job.id));
    return result.valueOrNull;
  }

  Future<DeliveryJob?> confirmPickup({
    required DeliveryJob job,
    required String scannedToken,
  }) async {
    final expected = job.handoffToken ?? scannedToken;
    final validation = DeliveryRules.validatePickupScan(
      status: job.status,
      scannedToken: scannedToken,
      expectedHandoffToken: expected,
    );
    if (validation.isFailure) return null;
    final result = await _repo.confirmPickup(
      id: job.id,
      handoffToken: scannedToken,
    );
    _ref.invalidate(activeDeliveriesProvider);
    _ref.invalidate(deliveryDetailProvider(job.id));
    return result.valueOrNull;
  }

  Future<DeliveryJob?> submitPod({
    required DeliveryJob job,
    required String photoPath,
    String? otp,
    String? signatureNote,
  }) async {
    final validation = DeliveryRules.validatePod(
      status: job.status,
      hasPhoto: photoPath.isNotEmpty,
      otp: otp,
      otpRequired: job.otpRequired,
    );
    if (validation.isFailure) return null;
    final result = await _repo.submitPod(
      id: job.id,
      photoPath: photoPath,
      otp: otp,
      signatureNote: signatureNote,
    );
    _ref.invalidate(activeDeliveriesProvider);
    _ref.invalidate(deliveryDetailProvider(job.id));
    return result.valueOrNull;
  }

  Future<DeliveryJob?> markFailed({
    required String id,
    required String reasonCode,
    String? note,
  }) async {
    final validation = DeliveryRules.validateFailureReason(reasonCode);
    if (validation.isFailure) return null;
    final result = await _repo.markFailed(
      id: id,
      reasonCode: reasonCode,
      note: note,
    );
    _ref.invalidate(activeDeliveriesProvider);
    _ref.invalidate(deliveryDetailProvider(id));
    return result.valueOrNull;
  }
}
