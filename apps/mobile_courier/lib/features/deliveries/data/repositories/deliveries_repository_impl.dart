import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/delivery_job.dart';
import '../../domain/repositories/deliveries_repository.dart';
import '../datasources/deliveries_remote_datasource.dart';

class DeliveriesRepositoryImpl implements DeliveriesRepository {
  DeliveriesRepositoryImpl(this._remote);
  final DeliveriesRemoteDataSource _remote;

  @override
  Future<Result<List<DeliveryJob>>> listActive() => _remote.listActive();

  @override
  Future<Result<DeliveryJob>> getById(String id) => _remote.getById(id);

  @override
  Future<Result<DeliveryJob>> transition(
    String id,
    DeliveryLifecycleStatus status,
  ) =>
      _remote.transition(id, status);

  @override
  Future<Result<DeliveryJob>> confirmPickup({
    required String id,
    required String handoffToken,
  }) =>
      _remote.confirmPickup(id: id, handoffToken: handoffToken);

  @override
  Future<Result<DeliveryJob>> submitPod({
    required String id,
    required String photoPath,
    String? otp,
    String? signatureNote,
  }) =>
      _remote.submitPod(
        id: id,
        photoPath: photoPath,
        otp: otp,
        signatureNote: signatureNote,
      );

  @override
  Future<Result<DeliveryJob>> markFailed({
    required String id,
    required String reasonCode,
    String? note,
  }) =>
      _remote.markFailed(id: id, reasonCode: reasonCode, note: note);
}
