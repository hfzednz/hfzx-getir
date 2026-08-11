import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/support_entity.dart';
import '../../domain/repositories/support_repository.dart';
import '../datasources/support_remote_datasource.dart';

class SupportRepositoryImpl implements SupportRepository {
  SupportRepositoryImpl(this._remote);
  final SupportRemoteDataSource _remote;

  @override
  Future<Result<List<SupportTicket>>> listTickets() => _remote.listTickets();

  @override
  Future<Result<List<ChatMessage>>> chatHistory() => _remote.chatHistory();

  @override
  Future<Result<ChatMessage>> sendChat(String message) =>
      _remote.sendChat(message);

  @override
  Future<Result<void>> reportIncident({
    required String type,
    required String description,
    double? lat,
    double? lng,
  }) =>
      _remote.reportIncident(
        type: type,
        description: description,
        lat: lat,
        lng: lng,
      );
}
