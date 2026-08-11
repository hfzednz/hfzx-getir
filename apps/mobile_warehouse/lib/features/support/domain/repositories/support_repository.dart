import 'package:nexora_core/nexora_core.dart';
import '../entities/support_entity.dart';

abstract class SupportRepository {
  Future<Result<List<SupportTicket>>> listTickets();
  Future<Result<SupportTicket>> createTicket({required String subject, required String body, required String idempotencyKey});
}
