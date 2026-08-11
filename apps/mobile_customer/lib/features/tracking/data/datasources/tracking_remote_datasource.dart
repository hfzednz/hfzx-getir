import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/tracking_entity.dart';
import '../models/tracking_model.dart';

class TrackingRemoteDataSource {
  const TrackingRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<TrackingSnapshot>> fetchSnapshot(String orderId) async {
    return _client.get<TrackingSnapshot>(
      '/orders/$orderId/tracking',
      parser: (json) =>
          TrackingSnapshotModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
