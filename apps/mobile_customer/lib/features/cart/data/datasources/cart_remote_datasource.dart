import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/cart_entity.dart';
import '../models/cart_model.dart';

class CartRemoteDataSource {
  const CartRemoteDataSource(this._client);
  final ApiClient _client;

  static const _basePath = '/cart';

  Cart _parseCart(dynamic json) =>
      CartModel.fromJson(json as Map<String, dynamic>).toEntity();

  Future<Result<Cart>> fetch({String? id}) async {
    final path = id != null ? '$_basePath/$id' : _basePath;
    return _client.get<Cart>(path, parser: (json) => _parseCart(json));
  }

  Future<Result<List<Cart>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Cart>>(
      _basePath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => CartModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<Cart>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Cart>(
      '$_basePath/items',
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> syncToCloud({required Map<String, dynamic> body}) async {
    return _client.post<Cart>(
      '$_basePath/sync',
      data: body,
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> mergeAnonymousCart({required String anonymousCartId}) async {
    return _client.post<Cart>(
      '$_basePath/merge',
      data: {'anonymous_cart_id': anonymousCartId},
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> validateInventory() async {
    return _client.post<Cart>(
      '$_basePath/validate',
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> applyCoupon(String code) async {
    return _client.post<Cart>(
      '$_basePath/coupon',
      data: {'code': code},
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> removeCoupon() async {
    return _client.delete<Cart>(
      '$_basePath/coupon',
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> applyGiftCard(String code) async {
    return _client.post<Cart>(
      '$_basePath/gift-card',
      data: {'code': code},
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> applyWallet(int amountMinor) async {
    return _client.post<Cart>(
      '$_basePath/wallet',
      data: {'amount_minor': amountMinor},
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> applyLoyaltyPoints(int points) async {
    return _client.post<Cart>(
      '$_basePath/loyalty',
      data: {'points': points},
      parser: (json) => _parseCart(json),
    );
  }

  Future<Result<Cart>> estimate() async {
    return _client.get<Cart>(
      '$_basePath/estimate',
      parser: (json) => _parseCart(json),
    );
  }
}
