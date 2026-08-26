import 'package:nexora_core/nexora_core.dart';

import '../../domain/checkout_bff_defaults.dart';
import '../../domain/entities/checkout_entity.dart';
import '../models/checkout_model.dart';

class CheckoutRemoteDataSource {
  const CheckoutRemoteDataSource(this._client);
  final ApiClient _client;

  static const _previewPath = '/checkout/preview';
  static const _placePath = '/checkout/place';
  static const _ordersPath = '/orders';

  CheckoutSession _parseSession(dynamic json) =>
      CheckoutSessionModel.fromJson(json as Map<String, dynamic>).toEntity();

  Map<String, dynamic> _bffBody(Map<String, dynamic> body) {
    final payment = body['payment'];
    final paymentType = body['paymentMethod'] ??
        (payment is Map ? payment['type'] : null) ??
        'card';
    return {
      'cartId': (body['cartId'] ?? body['cart_id'] ?? CheckoutBffDefaults.cartId).toString(),
      'principalId':
          (body['principalId'] ?? body['principal_id'] ?? CheckoutBffDefaults.principalId)
              .toString(),
      'paymentMethod': paymentType.toString(),
      if (body['sessionId'] != null || body['session_id'] != null)
        'sessionId': (body['sessionId'] ?? body['session_id']).toString(),
    };
  }

  Future<Result<CheckoutSession>> fetch({String? id}) async {
    if (id == null || id.isEmpty) {
      return const Failure<CheckoutSession>(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Order id required',
        ),
      );
    }
    return _client.get<CheckoutSession>(
      '$_ordersPath/$id',
      parser: _parseSession,
    );
  }

  Future<Result<List<CheckoutSession>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<CheckoutSession>>(
      _ordersPath,
      queryParameters: params,
      parser: (json) {
        final items = json is Map
            ? (json['items'] ?? json['Items'] ?? json['orders'] ?? [])
            : json;
        if (items is! List) return <CheckoutSession>[];
        return items
            .whereType<Map>()
            .map((e) => CheckoutSession.fromJson(Map<String, dynamic>.from(e)))
            .toList();
      },
    );
  }

  Future<Result<CheckoutSession>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<CheckoutSession>(
      _placePath,
      data: _bffBody(body),
      idempotencyKey: idempotencyKey,
      parser: (json) {
        final map = Map<String, dynamic>.from(json as Map);
        final orderId = map['orderId']?.toString() ?? map['order_id']?.toString();
        if (orderId != null && orderId.isNotEmpty) {
          return CheckoutSession(
            id: orderId,
            orderId: orderId,
            status: 'placed',
          );
        }
        return _parseSession(map);
      },
    );
  }

  Future<Result<CheckoutQuote>> getQuote({required Map<String, dynamic> body}) async {
    return _client.post<CheckoutQuote>(
      _previewPath,
      data: _bffBody(body),
      parser: (json) =>
          CheckoutQuote.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<CheckoutSession>> verifyQuote({
    required String quoteId,
    required Map<String, dynamic> body,
  }) async {
    return _client.post<CheckoutSession>(
      _previewPath,
      data: _bffBody(body),
      parser: (json) {
        final map = Map<String, dynamic>.from(json as Map);
        final quote = CheckoutQuote.fromJson({...map, 'quote_id': quoteId});
        return CheckoutSession(
          id: quote.quoteId ?? quoteId,
          status: 'quoted',
          quote: quote,
        );
      },
    );
  }
}
