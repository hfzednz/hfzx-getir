import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/notification_entity.dart';

class NotificationsRemoteDataSource {
  const NotificationsRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<List<WarehouseNotification>>> list() {
    return _client.get<List<WarehouseNotification>>(
      '/warehouse/notifications',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => WarehouseNotification.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }
  Future<Result<void>> markRead(String id) {
    return _client.post<void>('/warehouse/notifications/$id/read');
  }
}
