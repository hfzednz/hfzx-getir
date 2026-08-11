import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/checkout_entity.dart';
import '../models/checkout_model.dart';

class CheckoutRemoteDataSource {
  const CheckoutRemoteDataSource(this._client);
  final ApiClient _client;

  static const _sessionsPath = '/checkout/sessions';
  static const _confirmPath = '/checkout/confirm';
  static const _quotePath = '/checkout/quote';

  CheckoutSession _parseSession(dynamic json) =>
      CheckoutSessionModel.fromJson(json as Map<String, dynamic>).toEntity();

  Future<Result<CheckoutSession>> fetch({String? id}) async {
    final path = id != null ? '$_sessionsPath/$id' : _sessionsPath;
    return _client.get<CheckoutSession>(path, parser: (json) => _parseSession(json));
  }

  Future<Result<List<CheckoutSession>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<CheckoutSession>>(
      _sessionsPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => CheckoutSessionModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<CheckoutSession>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<CheckoutSession>(
      _confirmPath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => _parseSession(json),
    );
  }

  Future<Result<CheckoutQuote>> getQuote({required Map<String, dynamic> body}) async {
    return _client.post<CheckoutQuote>(
      _quotePath,
      data: body,
      parser: (json) => CheckoutQuoteModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<CheckoutSession>> verifyQuote({
    required String quoteId,
    required Map<String, dynamic> body,
  }) async {
    return _client.post<CheckoutSession>(
      '$_quotePath/$quoteId/verify',
      data: body,
      parser: (json) => _parseSession(json),
    );
  }
}
