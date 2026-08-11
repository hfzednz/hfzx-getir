import 'package:nexora_core/nexora_core.dart';

import '../entities/notification_entity.dart';

abstract class NotificationsRepository {
  Future<Result<List<CourierNotification>>> list();
  Future<Result<void>> markRead(String id);
}
