import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/tracking_entity.dart';
import '../../domain/repositories/tracking_repository.dart';
import '../datasources/tracking_remote_datasource.dart';

class TrackingRepositoryImpl implements TrackingRepository {
  const TrackingRepositoryImpl(this._remote);
  final TrackingRemoteDataSource _remote;

  @override
  Future<Result<TrackingSnapshot>> fetchSnapshot(String orderId) =>
      _remote.fetchSnapshot(orderId);

  @override
  Future<Result<RealtimeTicket>> issueRealtimeTicket(String orderId) =>
      _remote.issueRealtimeTicket(orderId);
}
