import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/notifications_remote_datasource.dart';
import '../../data/repositories/notifications_repository_impl.dart';
import '../../domain/entities/notification_entity.dart';
import '../../domain/repositories/notifications_repository.dart';

final notificationsRemoteDataSourceProvider = Provider((ref) => NotificationsRemoteDataSource(ref.watch(apiClientProvider)));
final notificationsRepositoryProvider = Provider<NotificationsRepository>((ref) => NotificationsRepositoryImpl(ref.watch(notificationsRemoteDataSourceProvider)));
final notificationsProvider = FutureProvider.autoDispose<List<WarehouseNotification>>((ref) async {
  final r = await ref.watch(notificationsRepositoryProvider).list();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
