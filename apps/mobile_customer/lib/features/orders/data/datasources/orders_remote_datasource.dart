import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/orders_entity.dart';
import '../models/orders_model.dart';

class OrdersRemoteDataSource {
  const OrdersRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/orders';

  Future<Result<Order>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<Order>(
      path,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<Order>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Order>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => OrderModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Order>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Order>(
      _listPath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Order>> cancelOrder({
    required String id,
    String? reason,
    String? idempotencyKey,
  }) async {
    return _client.post<Order>(
      '$_listPath/$id/cancel',
      data: ifNonEmpty({'reason': reason}),
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Order>> partialCancel({
    required String id,
    required List<String> lineIds,
    String? idempotencyKey,
  }) async {
    return _client.post<Order>(
      '$_listPath/$id/partial-cancel',
      data: {'line_ids': lineIds},
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Order>> requestRefund({
    required String id,
    String? reason,
    String? idempotencyKey,
  }) async {
    return _client.post<Order>(
      '$_listPath/$id/refund',
      data: ifNonEmpty({'reason': reason}),
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Order>> reorder({
    required String id,
    String? idempotencyKey,
  }) async {
    return _client.post<Order>(
      '$_listPath/$id/reorder',
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Order>> markFavoriteOrder(String id, {required bool favorite}) async {
    return _client.post<Order>(
      '$_listPath/$id/favorite',
      data: {'favorite': favorite},
      parser: (json) =>
          OrderModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<OrderDocument>> getInvoice(String id) async {
    return _client.get<OrderDocument>(
      '$_listPath/$id/invoice',
      parser: (json) => OrderDocument.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Result<OrderDocument>> getReceipt(String id) async {
    return _client.get<OrderDocument>(
      '$_listPath/$id/receipt',
      parser: (json) => OrderDocument.fromJson(json as Map<String, dynamic>),
    );
  }

  Future<Result<String>> fetchRealtimeTicket(String id) async {
    return _client.get<String>(
      '$_listPath/$id/realtime-ticket',
      parser: (json) {
        final map = json as Map<String, dynamic>;
        return map['ticket']?.toString() ?? map['token']?.toString() ?? '';
      },
    );
  }
}

Map<String, dynamic> ifNonEmpty(Map<String, String?> fields) {
  return {
    for (final entry in fields.entries)
      if (entry.value != null && entry.value!.isNotEmpty) entry.key: entry.value,
  };
}
