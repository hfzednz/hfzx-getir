import 'package:nexora_core/nexora_core.dart';

import '../entities/tracking_entity.dart';

abstract class TrackingRepository {
  Future<Result<TrackingSnapshot>> fetchSnapshot(String orderId);
  Future<Result<String>> issueRealtimeTicket(String orderId);
}
