import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/product_entity.dart';
import '../models/product_model.dart';

class ProductRemoteDataSource {
  const ProductRemoteDataSource(this._client);
  final ApiClient _client;

  static const _basePath = '/products';

  Future<Result<Product>> fetch({String? id}) async {
    final path = id != null ? '$_basePath/$id' : _basePath;
    return _client.get<Product>(
      path,
      parser: (json) => ProductModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Product>> getProduct(String id) => fetch(id: id);

  Future<Result<List<Product>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Product>>(
      _basePath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => ProductModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Product>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Product>(
      _basePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => ProductModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<ProductSummary>>> getAlternatives(String productId) async {
    return _client.get<List<ProductSummary>>(
      '$_basePath/$productId/alternatives',
      parser: (json) => (json['items'] as List<dynamic>? ?? json as List<dynamic>)
          .map((e) => ProductSummary.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  Future<Result<List<ProductPricePoint>>> getPriceHistory(String productId) async {
    return _client.get<List<ProductPricePoint>>(
      '$_basePath/$productId/price-history',
      parser: (json) => ProductPriceHistoryModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<void>> askQuestion({
    required String productId,
    required String question,
  }) async {
    return _client.post<void>(
      '$_basePath/$productId/questions',
      data: {'question': question},
    );
  }

  Future<Result<bool>> toggleFavorite(String productId) async {
    return _client.post<bool>(
      '$_basePath/$productId/favorite',
      parser: (json) => (json as Map<String, dynamic>)['is_favorite'] == true,
    );
  }
}
