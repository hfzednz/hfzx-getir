import 'package:nexora_core/nexora_core.dart';

import '../entities/notifications_entity.dart';
import '../repositories/notifications_repository.dart';

class GetNotificationsUseCase {
  const GetNotificationsUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<AppNotification>> call({String? id}) => _repository.fetch(id: id);
}

class ListNotificationsUseCase {
  const ListNotificationsUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<List<AppNotification>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class MarkNotificationReadUseCase {
  const MarkNotificationReadUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<AppNotification>> call(String id) => _repository.markRead(id);
}

class MarkAllNotificationsReadUseCase {
  const MarkAllNotificationsReadUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<void>> call() => _repository.markAllRead();
}

class RegisterFcmTokenUseCase {
  const RegisterFcmTokenUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<void>> call(String token) => _repository.registerFcmToken(token);
}

class GetNotificationPreferencesUseCase {
  const GetNotificationPreferencesUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<NotificationPreferences>> call() => _repository.fetchPreferences();
}

class UpdateNotificationPreferencesUseCase {
  const UpdateNotificationPreferencesUseCase(this._repository);
  final NotificationsRepository _repository;

  Future<Result<NotificationPreferences>> call(NotificationPreferences prefs) =>
      _repository.updatePreferences(prefs);
}
