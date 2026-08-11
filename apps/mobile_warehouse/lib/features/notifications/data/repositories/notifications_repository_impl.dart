import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/notification_entity.dart';
import '../../domain/repositories/notifications_repository.dart';
import '../datasources/notifications_remote_datasource.dart';

class NotificationsRepositoryImpl implements NotificationsRepository {
  NotificationsRepositoryImpl(this._remote);
  final NotificationsRemoteDataSource _remote;
  @override
  Future<Result<List<WarehouseNotification>>> list() => _remote.list();
  @override
  Future<Result<void>> markRead(String id) => _remote.markRead(id);
}
