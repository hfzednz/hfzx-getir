import 'package:nexora_core/nexora_core.dart';

import '../entities/product_entity.dart';

abstract class ProductRepository {
  Future<Result<Product>> fetch({String? id});
  Future<Result<List<Product>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Product>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<Product>> getProduct(String id);
  Future<Result<List<ProductSummary>>> getAlternatives(String productId);
  Future<Result<List<ProductPricePoint>>> getPriceHistory(String productId);
  Future<Result<void>> askQuestion({required String productId, required String question});
  Future<Result<bool>> toggleFavorite(String productId);
}
