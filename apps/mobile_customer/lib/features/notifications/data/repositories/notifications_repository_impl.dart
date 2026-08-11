import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/notifications_entity.dart';
import '../../domain/repositories/notifications_repository.dart';
import '../datasources/notifications_remote_datasource.dart';

class NotificationsRepositoryImpl implements NotificationsRepository {
  const NotificationsRepositoryImpl(this._remote);
  final NotificationsRemoteDataSource _remote;

  @override
  Future<Result<AppNotification>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<AppNotification>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<AppNotification>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<AppNotification>> markRead(String id) => _remote.markRead(id);

  @override
  Future<Result<void>> markAllRead() => _remote.markAllRead();

  @override
  Future<Result<void>> registerFcmToken(String token) =>
      _remote.registerFcmToken(token);

  @override
  Future<Result<NotificationPreferences>> fetchPreferences() =>
      _remote.fetchPreferences();

  @override
  Future<Result<NotificationPreferences>> updatePreferences(
    NotificationPreferences preferences,
  ) =>
      _remote.updatePreferences(preferences);
}
