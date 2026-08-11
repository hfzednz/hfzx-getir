import 'package:nexora_core/nexora_core.dart';

import '../entities/checkout_entity.dart';
import '../repositories/checkout_repository.dart';

class GetCheckoutUseCase {
  const GetCheckoutUseCase(this._repository);
  final CheckoutRepository _repository;

  Future<Result<CheckoutSession>> call({String? id}) => _repository.fetch(id: id);
}

class ListCheckoutUseCase {
  const ListCheckoutUseCase(this._repository);
  final CheckoutRepository _repository;

  Future<Result<List<CheckoutSession>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class GetCheckoutQuoteUseCase {
  const GetCheckoutQuoteUseCase(this._repository);
  final CheckoutRepository _repository;

  Future<Result<CheckoutQuote>> call({required Map<String, dynamic> body}) =>
      _repository.getQuote(body: body);
}

class ConfirmCheckoutUseCase {
  const ConfirmCheckoutUseCase(this._repository);
  final CheckoutRepository _repository;

  Future<Result<CheckoutSession>> call({
    required Map<String, dynamic> body,
    required String idempotencyKey,
  }) =>
      _repository.mutate(body: body, idempotencyKey: idempotencyKey);
}

class VerifyCheckoutQuoteUseCase {
  const VerifyCheckoutQuoteUseCase(this._repository);
  final CheckoutRepository _repository;

  Future<Result<CheckoutSession>> call({
    required String quoteId,
    required Map<String, dynamic> body,
  }) =>
      _repository.verifyQuote(quoteId: quoteId, body: body);
}
