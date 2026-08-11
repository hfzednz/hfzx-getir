import 'package:nexora_core/nexora_core.dart';

import '../entities/support_entity.dart';
import '../repositories/support_repository.dart';

class ListSupportFaqUseCase {
  const ListSupportFaqUseCase(this._repository);
  final SupportRepository _repository;

  Future<Result<List<SupportFaq>>> call() => _repository.fetchFaq();
}

class ListSupportTicketsUseCase {
  const ListSupportTicketsUseCase(this._repository);
  final SupportRepository _repository;

  Future<Result<List<SupportTicket>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchTickets(params: params);
}

class CreateSupportTicketUseCase {
  const CreateSupportTicketUseCase(this._repository);
  final SupportRepository _repository;

  Future<Result<SupportTicket>> call({
    required String subject,
    required String body,
    String? orderId,
    String? category,
    String? idempotencyKey,
  }) =>
      _repository.createTicket(
        subject: subject,
        body: body,
        orderId: orderId,
        category: category,
        idempotencyKey: idempotencyKey,
      );
}
