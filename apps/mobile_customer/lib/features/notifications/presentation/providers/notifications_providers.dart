import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/notifications_remote_datasource.dart';
import '../../data/repositories/notifications_repository_impl.dart';
import '../../domain/entities/notifications_entity.dart';
import '../../domain/repositories/notifications_repository.dart';
import '../../domain/usecases/notifications_usecases.dart';

final notificationsRemoteDataSourceProvider =
    Provider<NotificationsRemoteDataSource>((ref) {
  return NotificationsRemoteDataSource(ref.watch(apiClientProvider));
});

final notificationsRepositoryProvider = Provider<NotificationsRepository>((ref) {
  return NotificationsRepositoryImpl(ref.watch(notificationsRemoteDataSourceProvider));
});

final getNotificationsUseCaseProvider = Provider(
  (ref) => GetNotificationsUseCase(ref.watch(notificationsRepositoryProvider)),
);

final listNotificationsUseCaseProvider = Provider(
  (ref) => ListNotificationsUseCase(ref.watch(notificationsRepositoryProvider)),
);

final markNotificationReadUseCaseProvider = Provider(
  (ref) => MarkNotificationReadUseCase(ref.watch(notificationsRepositoryProvider)),
);

final markAllNotificationsReadUseCaseProvider = Provider(
  (ref) => MarkAllNotificationsReadUseCase(ref.watch(notificationsRepositoryProvider)),
);

final registerFcmTokenUseCaseProvider = Provider(
  (ref) => RegisterFcmTokenUseCase(ref.watch(notificationsRepositoryProvider)),
);

final getNotificationPreferencesUseCaseProvider = Provider(
  (ref) => GetNotificationPreferencesUseCase(ref.watch(notificationsRepositoryProvider)),
);

final updateNotificationPreferencesUseCaseProvider = Provider(
  (ref) =>
      UpdateNotificationPreferencesUseCase(ref.watch(notificationsRepositoryProvider)),
);

final notificationsListProvider =
    FutureProvider.autoDispose<List<AppNotification>>((ref) async {
  final result = await ref.watch(listNotificationsUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});

final notificationPreferencesProvider =
    FutureProvider.autoDispose<NotificationPreferences>((ref) async {
  final result = await ref.watch(getNotificationPreferencesUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final unreadNotificationsCountProvider = Provider<int>((ref) {
  final asyncList = ref.watch(notificationsListProvider);
  return asyncList.maybeWhen(
    data: (items) => items.where((n) => !n.read).length,
    orElse: () => 0,
  );
});
