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
  Future<Result<SupportTicket>> createTicket({required String subject, required String body, required String idempotencyKey}) =>
      _remote.createTicket(subject: subject, body: body, idempotencyKey: idempotencyKey);
}
