import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/support_entity.dart';

class SupportRemoteDataSource {
  const SupportRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<List<SupportTicket>>> listTickets() {
    return _client.get<List<SupportTicket>>(
      '/warehouse/support/tickets',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => SupportTicket.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }
  Future<Result<SupportTicket>> createTicket({required String subject, required String body, required String idempotencyKey}) {
    return _client.post<SupportTicket>(
      '/warehouse/support/tickets',
      data: {'subject': subject, 'body': body},
      idempotencyKey: idempotencyKey,
      parser: (json) => SupportTicket.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
