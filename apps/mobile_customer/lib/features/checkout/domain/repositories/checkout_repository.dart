import 'package:nexora_core/nexora_core.dart';

import '../entities/checkout_entity.dart';

abstract class CheckoutRepository {
  Future<Result<CheckoutSession>> fetch({String? id});
  Future<Result<List<CheckoutSession>>> fetchList({Map<String, dynamic>? params});
  Future<Result<CheckoutSession>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<CheckoutQuote>> getQuote({required Map<String, dynamic> body});
  Future<Result<CheckoutSession>> verifyQuote({
    required String quoteId,
    required Map<String, dynamic> body,
  });
}

abstract class PaymentMethodsRepository {
  Future<Result<List<SavedPaymentCard>>> listSavedCards();
  Future<Result<CheckoutSession>> payWithWallet({required Map<String, dynamic> body});
  Future<Result<CheckoutSession>> retryPayment({required String sessionId});
}
