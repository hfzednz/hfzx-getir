import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/orders_entity.dart';
import '../../domain/repositories/orders_repository.dart';
import '../datasources/orders_remote_datasource.dart';

class OrdersRepositoryImpl implements OrdersRepository {
  const OrdersRepositoryImpl(this._remote);
  final OrdersRemoteDataSource _remote;

  @override
  Future<Result<Order>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<Order>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Order>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Order>> cancelOrder({
    required String id,
    String? reason,
    String? idempotencyKey,
  }) =>
      _remote.cancelOrder(id: id, reason: reason, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Order>> partialCancel({
    required String id,
    required List<String> lineIds,
    String? idempotencyKey,
  }) =>
      _remote.partialCancel(
        id: id,
        lineIds: lineIds,
        idempotencyKey: idempotencyKey,
      );

  @override
  Future<Result<Order>> requestRefund({
    required String id,
    String? reason,
    String? idempotencyKey,
  }) =>
      _remote.requestRefund(id: id, reason: reason, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Order>> reorder({
    required String id,
    String? idempotencyKey,
  }) =>
      _remote.reorder(id: id, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Order>> markFavoriteOrder(String id, {required bool favorite}) =>
      _remote.markFavoriteOrder(id, favorite: favorite);

  @override
  Future<Result<OrderDocument>> getInvoice(String id) => _remote.getInvoice(id);

  @override
  Future<Result<OrderDocument>> getReceipt(String id) => _remote.getReceipt(id);

  @override
  Future<Result<String>> fetchRealtimeTicket(String id) =>
      _remote.fetchRealtimeTicket(id);
}
