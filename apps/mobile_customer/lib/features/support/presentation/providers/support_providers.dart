import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../data/datasources/support_remote_datasource.dart';
import '../../data/repositories/support_repository_impl.dart';
import '../../domain/entities/support_entity.dart';
import '../../domain/repositories/support_repository.dart';

final supportRemoteDataSourceProvider = Provider<SupportRemoteDataSource>((ref) {
  return SupportRemoteDataSource(ref.watch(apiClientProvider));
});

final supportRepositoryProvider = Provider<SupportRepository>((ref) {
  return SupportRepositoryImpl(ref.watch(supportRemoteDataSourceProvider));
});

final supportFaqProvider = FutureProvider.autoDispose<List<SupportFaq>>((ref) async {
  final result = await ref.watch(supportRepositoryProvider).fetchFaq();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final supportTicketsProvider = FutureProvider.autoDispose<List<SupportTicket>>((ref) async {
  final result = await ref.watch(supportRepositoryProvider).fetchTickets();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final supportTicketProvider =
    FutureProvider.autoDispose.family<SupportTicket, String>((ref, id) async {
  final result = await ref.watch(supportRepositoryProvider).fetchTicket(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final supportAssistantMessagesProvider =
    StateProvider<List<SupportAssistantMessage>>((ref) => const []);

final supportTicketCreateProvider =
    AsyncNotifierProvider<SupportTicketCreateController, SupportTicket?>(SupportTicketCreateController.new);

class SupportTicketCreateController extends AsyncNotifier<SupportTicket?> {
  @override
  Future<SupportTicket?> build() async => null;

  Future<void> create({
    required String subject,
    required String body,
    String? orderId,
    String? category,
  }) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final result = await ref.read(supportRepositoryProvider).createTicket(
            subject: subject,
            body: body,
            orderId: orderId,
            category: category,
            idempotencyKey: const Uuid().v4(),
          );
      return result.fold(
        onSuccess: (ticket) {
          ref.invalidate(supportTicketsProvider);
          ref.read(analyticsTrackerProvider).trackRaw(
                eventName: AnalyticsEvents.supportTicketCreated,
                props: {
                  'ticket_id': ticket.id,
                  if (orderId != null) 'order_id': orderId,
                },
              );
          return ticket;
        },
        onFailure: (e) => throw e,
      );
    });
  }
}

final supportAssistantSendProvider =
    AsyncNotifierProvider<SupportAssistantSendController, void>(SupportAssistantSendController.new);

class SupportAssistantSendController extends AsyncNotifier<void> {
  @override
  Future<void> build() async {}

  Future<void> send(String message, {String? orderId}) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final userMsg = SupportAssistantMessage(id: DateTime.now().millisecondsSinceEpoch.toString(), role: 'user', content: message);
      ref.read(supportAssistantMessagesProvider.notifier).update((s) => [...s, userMsg]);

      final result = await ref.read(supportRepositoryProvider).sendAssistantMessage(
            message: message,
            orderId: orderId,
          );
      result.fold(
        onSuccess: (reply) {
          ref.read(supportAssistantMessagesProvider.notifier).update((s) => [...s, reply]);
        },
        onFailure: (e) => throw e,
      );
    });
  }
}
