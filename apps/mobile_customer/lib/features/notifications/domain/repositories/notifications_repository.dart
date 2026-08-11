import 'package:nexora_core/nexora_core.dart';

import '../entities/notifications_entity.dart';

abstract class NotificationsRepository {
  Future<Result<AppNotification>> fetch({String? id});
  Future<Result<List<AppNotification>>> fetchList({Map<String, dynamic>? params});
  Future<Result<AppNotification>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<AppNotification>> markRead(String id);

  Future<Result<void>> markAllRead();

  Future<Result<void>> registerFcmToken(String token);

  Future<Result<NotificationPreferences>> fetchPreferences();

  Future<Result<NotificationPreferences>> updatePreferences(
    NotificationPreferences preferences,
  );
}
