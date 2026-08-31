import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/tracking_entity.dart';
import '../models/tracking_model.dart';

class TrackingRemoteDataSource {
  const TrackingRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<TrackingSnapshot>> fetchSnapshot(String orderId) async {
    return _client.get<TrackingSnapshot>(
      '/orders/$orderId/track',
      parser: (json) =>
          TrackingSnapshotModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<String>> issueRealtimeTicket(String orderId) async {
    return _client.post<String>(
      '/orders/$orderId/realtime-ticket',
      parser: (json) {
        final map = json as Map<String, dynamic>;
        return map['ticket']?.toString() ?? map['token']?.toString() ?? '';
      },
    );
  }
}
