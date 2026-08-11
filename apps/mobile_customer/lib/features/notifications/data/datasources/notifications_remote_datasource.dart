import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/notifications_entity.dart';
import '../models/notifications_model.dart';

class NotificationsRemoteDataSource {
  const NotificationsRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/notifications';
  static const _prefsPath = '/notifications/preferences';
  static const _fcmPath = '/notifications/fcm-token';

  Future<Result<AppNotification>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<AppNotification>(
      path,
      parser: (json) =>
          AppNotificationModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<AppNotification>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<AppNotification>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map(
            (e) =>
                AppNotificationModel.fromJson(e as Map<String, dynamic>).toEntity(),
          )
          .toList(),
    );
  }

  Future<Result<AppNotification>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<AppNotification>(
      _listPath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          AppNotificationModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<AppNotification>> markRead(String id) async {
    return _client.post<AppNotification>(
      '$_listPath/$id/read',
      parser: (json) =>
          AppNotificationModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<void>> markAllRead() async {
    return _client.post<void>(
      '$_listPath/read-all',
      parser: (_) {},
    );
  }

  Future<Result<void>> registerFcmToken(String token) async {
    return _client.post<void>(
      _fcmPath,
      data: {'token': token, 'platform': 'mobile'},
      parser: (_) {},
    );
  }

  Future<Result<NotificationPreferences>> fetchPreferences() async {
    return _client.get<NotificationPreferences>(
      _prefsPath,
      parser: (json) =>
          NotificationPreferences.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Result<NotificationPreferences>> updatePreferences(
    NotificationPreferences preferences,
  ) async {
    return _client.put<NotificationPreferences>(
      _prefsPath,
      data: preferences.toJson(),
      parser: (json) =>
          NotificationPreferences.fromJson(json as Map<String, dynamic>),
    );
  }
}
