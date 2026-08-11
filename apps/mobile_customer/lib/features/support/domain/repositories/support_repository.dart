import 'package:nexora_core/nexora_core.dart';

import '../entities/support_entity.dart';

abstract class SupportRepository {
  Future<Result<List<SupportFaq>>> fetchFaq();
  Future<Result<List<SupportTicket>>> fetchTickets({Map<String, dynamic>? params});
  Future<Result<SupportTicket>> fetchTicket(String id);
  Future<Result<SupportTicket>> createTicket({
    required String subject,
    required String body,
    String? orderId,
    String? category,
    String? idempotencyKey,
  });
  Future<Result<SupportAssistantMessage>> sendAssistantMessage({
    required String message,
    String? sessionId,
    String? orderId,
  });
}
