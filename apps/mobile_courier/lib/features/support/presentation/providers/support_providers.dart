import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/support_remote_datasource.dart';
import '../../data/repositories/support_repository_impl.dart';
import '../../domain/entities/support_entity.dart';
import '../../domain/repositories/support_repository.dart';

final supportRemoteDataSourceProvider =
    Provider<SupportRemoteDataSource>((ref) {
  return SupportRemoteDataSource(ref.watch(apiClientProvider));
});

final supportRepositoryProvider = Provider<SupportRepository>((ref) {
  return SupportRepositoryImpl(ref.watch(supportRemoteDataSourceProvider));
});

final supportTicketsProvider =
    FutureProvider.autoDispose<List<SupportTicket>>((ref) async {
  final result = await ref.watch(supportRepositoryProvider).listTickets();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final supportChatProvider =
    FutureProvider.autoDispose<List<ChatMessage>>((ref) async {
  final result = await ref.watch(supportRepositoryProvider).chatHistory();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final supportActionsProvider = Provider((ref) => SupportActions(ref));

class SupportActions {
  SupportActions(this._ref);
  final Ref _ref;

  Future<void> sendChat(String message) async {
    await _ref.read(supportRepositoryProvider).sendChat(message);
    _ref.invalidate(supportChatProvider);
  }

  Future<void> reportIncident({
    required String type,
    required String description,
  }) async {
    await _ref.read(supportRepositoryProvider).reportIncident(
          type: type,
          description: description,
        );
  }
}
