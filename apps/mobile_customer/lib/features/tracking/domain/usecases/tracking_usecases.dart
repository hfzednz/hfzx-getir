import 'package:nexora_core/nexora_core.dart';

import '../entities/tracking_entity.dart';
import '../repositories/tracking_repository.dart';

class GetTrackingSnapshotUseCase {
  const GetTrackingSnapshotUseCase(this._repository);
  final TrackingRepository _repository;

  Future<Result<TrackingSnapshot>> call(String orderId) =>
      _repository.fetchSnapshot(orderId);
}
