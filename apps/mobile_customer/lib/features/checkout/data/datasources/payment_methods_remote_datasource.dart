import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/checkout_entity.dart';
import '../models/checkout_model.dart';

class PaymentMethodsRemoteDataSource {
  const PaymentMethodsRemoteDataSource(this._client);
  final ApiClient _client;

  static const _cardsPath = '/payment-methods/cards';
  static const _walletPayPath = '/payment-methods/wallet/pay';
  static const _retryPath = '/payment-methods/retry';

  Future<Result<List<SavedPaymentCard>>> listSavedCards() async {
    return _client.get<List<SavedPaymentCard>>(
      _cardsPath,
      parser: (json) => (json['cards'] as List<dynamic>? ?? json as List<dynamic>)
          .map((e) => SavedPaymentCardModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<CheckoutSession>> payWithWallet({required Map<String, dynamic> body}) async {
    return _client.post<CheckoutSession>(
      _walletPayPath,
      data: body,
      parser: (json) =>
          CheckoutSessionModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<CheckoutSession>> retryPayment({required String sessionId}) async {
    return _client.post<CheckoutSession>(
      _retryPath,
      data: {'session_id': sessionId},
      parser: (json) =>
          CheckoutSessionModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
