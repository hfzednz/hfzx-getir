import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/checkout_entity.dart';
import '../../domain/repositories/checkout_repository.dart';
import '../datasources/checkout_remote_datasource.dart';
import '../datasources/payment_methods_remote_datasource.dart';

class CheckoutRepositoryImpl implements CheckoutRepository {
  const CheckoutRepositoryImpl(this._remote);
  final CheckoutRemoteDataSource _remote;

  @override
  Future<Result<CheckoutSession>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<CheckoutSession>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<CheckoutSession>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<CheckoutQuote>> getQuote({required Map<String, dynamic> body}) =>
      _remote.getQuote(body: body);

  @override
  Future<Result<CheckoutSession>> verifyQuote({
    required String quoteId,
    required Map<String, dynamic> body,
  }) =>
      _remote.verifyQuote(quoteId: quoteId, body: body);
}

class PaymentMethodsRepositoryImpl implements PaymentMethodsRepository {
  const PaymentMethodsRepositoryImpl(this._remote);
  final PaymentMethodsRemoteDataSource _remote;

  @override
  Future<Result<List<SavedPaymentCard>>> listSavedCards() => _remote.listSavedCards();

  @override
  Future<Result<CheckoutSession>> payWithWallet({required Map<String, dynamic> body}) =>
      _remote.payWithWallet(body: body);

  @override
  Future<Result<CheckoutSession>> retryPayment({required String sessionId}) =>
      _remote.retryPayment(sessionId: sessionId);
}
