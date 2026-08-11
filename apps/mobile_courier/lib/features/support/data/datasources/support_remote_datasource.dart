import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/support_entity.dart';

class SupportRemoteDataSource {
  const SupportRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<SupportTicket>>> listTickets() {
    return _client.get<List<SupportTicket>>(
      '/courier/support/tickets',
      parser: (json) {
        final list = json is List
            ? json
            : (json as Map)['tickets'] as List? ?? const [];
        return list
            .map((e) =>
                SupportTicket.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<List<ChatMessage>>> chatHistory() {
    return _client.get<List<ChatMessage>>(
      '/courier/support/assistant',
      parser: (json) {
        final list = json is List
            ? json
            : (json as Map)['messages'] as List? ?? const [];
        return list
            .map((e) =>
                ChatMessage.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<ChatMessage>> sendChat(String message) {
    return _client.post<ChatMessage>(
      '/courier/support/assistant',
      data: {'message': message},
      parser: (json) =>
          ChatMessage.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<void>> reportIncident({
    required String type,
    required String description,
    double? lat,
    double? lng,
  }) {
    return _client.post<void>(
      '/courier/incidents',
      data: {
        'type': type,
        'description': description,
        if (lat != null) 'lat': lat,
        if (lng != null) 'lng': lng,
      },
      parser: (_) {},
    );
  }
}
