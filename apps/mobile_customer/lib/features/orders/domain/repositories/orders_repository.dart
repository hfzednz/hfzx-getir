import 'package:nexora_core/nexora_core.dart';

import '../entities/orders_entity.dart';

abstract class OrdersRepository {
  Future<Result<Order>> fetch({String? id});
  Future<Result<List<Order>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Order>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<Order>> cancelOrder({
    required String id,
    String? reason,
    String? idempotencyKey,
  });

  Future<Result<Order>> partialCancel({
    required String id,
    required List<String> lineIds,
    String? idempotencyKey,
  });

  Future<Result<Order>> requestRefund({
    required String id,
    String? reason,
    String? idempotencyKey,
  });

  Future<Result<Order>> reorder({
    required String id,
    String? idempotencyKey,
  });

  Future<Result<Order>> markFavoriteOrder(String id, {required bool favorite});

  Future<Result<OrderDocument>> getInvoice(String id);

  Future<Result<OrderDocument>> getReceipt(String id);

  Future<Result<String>> fetchRealtimeTicket(String id);
}
