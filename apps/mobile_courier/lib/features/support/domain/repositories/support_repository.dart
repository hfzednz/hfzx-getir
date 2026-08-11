import 'package:nexora_core/nexora_core.dart';

import '../entities/support_entity.dart';

abstract class SupportRepository {
  Future<Result<List<SupportTicket>>> listTickets();
  Future<Result<List<ChatMessage>>> chatHistory();
  Future<Result<ChatMessage>> sendChat(String message);
  Future<Result<void>> reportIncident({
    required String type,
    required String description,
    double? lat,
    double? lng,
  });
}
