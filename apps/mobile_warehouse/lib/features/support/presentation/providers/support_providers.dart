import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/support_remote_datasource.dart';
import '../../data/repositories/support_repository_impl.dart';
import '../../domain/entities/support_entity.dart';
import '../../domain/repositories/support_repository.dart';

final supportRemoteDataSourceProvider = Provider((ref) => SupportRemoteDataSource(ref.watch(apiClientProvider)));
final supportRepositoryProvider = Provider<SupportRepository>((ref) => SupportRepositoryImpl(ref.watch(supportRemoteDataSourceProvider)));
final supportTicketsProvider = FutureProvider.autoDispose<List<SupportTicket>>((ref) async {
  final r = await ref.watch(supportRepositoryProvider).listTickets();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final supportActionsProvider = Provider((ref) => SupportActions(ref));

class SupportActions {
  SupportActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();
  Future<Result<SupportTicket>> create({required String subject, required String body}) async {
    final r = await _ref.read(supportRepositoryProvider).createTicket(subject: subject, body: body, idempotencyKey: _uuid.v4());
    if (r.isSuccess) _ref.invalidate(supportTicketsProvider);
    return r;
  }
}
