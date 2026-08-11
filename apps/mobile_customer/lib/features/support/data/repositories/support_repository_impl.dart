import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/support_entity.dart';
import '../../domain/repositories/support_repository.dart';
import '../datasources/support_remote_datasource.dart';

class SupportRepositoryImpl implements SupportRepository {
  const SupportRepositoryImpl(this._remote);
  final SupportRemoteDataSource _remote;

  @override
  Future<Result<List<SupportFaq>>> fetchFaq() => _remote.fetchFaq();

  @override
  Future<Result<List<SupportTicket>>> fetchTickets({Map<String, dynamic>? params}) =>
      _remote.fetchTickets(params: params);

  @override
  Future<Result<SupportTicket>> fetchTicket(String id) => _remote.fetchTicket(id);

  @override
  Future<Result<SupportTicket>> createTicket({
    required String subject,
    required String body,
    String? orderId,
    String? category,
    String? idempotencyKey,
  }) =>
      _remote.createTicket(
        subject: subject,
        body: body,
        orderId: orderId,
        category: category,
        idempotencyKey: idempotencyKey,
      );

  @override
  Future<Result<SupportAssistantMessage>> sendAssistantMessage({
    required String message,
    String? sessionId,
    String? orderId,
  }) =>
      _remote.sendAssistantMessage(message: message, sessionId: sessionId, orderId: orderId);
}
