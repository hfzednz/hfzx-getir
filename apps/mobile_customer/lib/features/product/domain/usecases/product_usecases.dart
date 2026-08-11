import 'package:nexora_core/nexora_core.dart';

import '../entities/product_entity.dart';
import '../repositories/product_repository.dart';

class GetProductUseCase {
  const GetProductUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<Product>> call({String? id}) =>
      id != null ? _repository.getProduct(id) : _repository.fetch();
}

class ListProductUseCase {
  const ListProductUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<List<Product>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class GetProductAlternativesUseCase {
  const GetProductAlternativesUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<List<ProductSummary>>> call(String productId) =>
      _repository.getAlternatives(productId);
}

class GetProductPriceHistoryUseCase {
  const GetProductPriceHistoryUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<List<ProductPricePoint>>> call(String productId) =>
      _repository.getPriceHistory(productId);
}

class AskProductQuestionUseCase {
  const AskProductQuestionUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<void>> call({required String productId, required String question}) =>
      _repository.askQuestion(productId: productId, question: question);
}

class ToggleProductFavoriteUseCase {
  const ToggleProductFavoriteUseCase(this._repository);
  final ProductRepository _repository;

  Future<Result<bool>> call(String productId) => _repository.toggleFavorite(productId);
}
