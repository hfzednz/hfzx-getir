import 'package:nexora_core/nexora_core.dart';

import '../entities/delivery_job.dart';

abstract class DeliveriesRepository {
  Future<Result<List<DeliveryJob>>> listActive();
  Future<Result<DeliveryJob>> getById(String id);
  Future<Result<DeliveryJob>> transition(
    String id,
    DeliveryLifecycleStatus status,
  );
  Future<Result<DeliveryJob>> confirmPickup({
    required String id,
    required String handoffToken,
  });
  Future<Result<DeliveryJob>> submitPod({
    required String id,
    required String photoPath,
    String? otp,
    String? signatureNote,
  });
  Future<Result<DeliveryJob>> markFailed({
    required String id,
    required String reasonCode,
    String? note,
  });
}
