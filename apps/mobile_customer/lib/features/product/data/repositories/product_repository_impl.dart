import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/product_entity.dart';
import '../../domain/repositories/product_repository.dart';
import '../datasources/product_remote_datasource.dart';

class ProductRepositoryImpl implements ProductRepository {
  const ProductRepositoryImpl(this._remote);
  final ProductRemoteDataSource _remote;

  @override
  Future<Result<Product>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<Product>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Product>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Product>> getProduct(String id) => _remote.getProduct(id);

  @override
  Future<Result<List<ProductSummary>>> getAlternatives(String productId) =>
      _remote.getAlternatives(productId);

  @override
  Future<Result<List<ProductPricePoint>>> getPriceHistory(String productId) =>
      _remote.getPriceHistory(productId);

  @override
  Future<Result<void>> askQuestion({
    required String productId,
    required String question,
  }) =>
      _remote.askQuestion(productId: productId, question: question);

  @override
  Future<Result<bool>> toggleFavorite(String productId) =>
      _remote.toggleFavorite(productId);
}
