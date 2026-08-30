import 'package:nexora_core/nexora_core.dart';

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

  static const _missingIdentity = NexoraValidationException(
    code: NexoraErrorCode.validationFailed,
    message: 'A cart and a signed-in customer are required for checkout',
  );

  static String? _id(Map<String, dynamic> body, String camel, String snake) {
    final raw = (body[camel] ?? body[snake])?.toString().trim();
    return raw == null || raw.isEmpty ? null : raw;
  }

  /// Maps the controller body onto the BFF contract. Returns null when the cart or the
  /// customer identity is unknown: checkout is keyed by both, and substituting a default
  /// would quote someone else's cart.
  Map<String, dynamic>? _bffBody(Map<String, dynamic> body) {
    final cartId = _id(body, 'cartId', 'cart_id');
    final principalId = _id(body, 'principalId', 'principal_id');
    if (cartId == null || principalId == null) return null;
    final payment = body['payment'];
    final paymentType = body['paymentMethod'] ??
        (payment is Map ? payment['type'] : null) ??
        'card';
    return {
      'cartId': cartId,
      'principalId': principalId,
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
    final data = _bffBody(body);
    if (data == null) {
      return const Failure<CheckoutSession>(_missingIdentity);
    }
    return _client.post<CheckoutSession>(
      _placePath,
      data: data,
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
    final data = _bffBody(body);
    if (data == null) {
      return const Failure<CheckoutQuote>(_missingIdentity);
    }
    return _client.post<CheckoutQuote>(
      _previewPath,
      data: data,
      parser: (json) =>
          CheckoutQuote.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<CheckoutSession>> verifyQuote({
    required String quoteId,
    required Map<String, dynamic> body,
  }) async {
    final data = _bffBody(body);
    if (data == null) {
      return const Failure<CheckoutSession>(_missingIdentity);
    }
    return _client.post<CheckoutSession>(
      _previewPath,
      data: data,
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
