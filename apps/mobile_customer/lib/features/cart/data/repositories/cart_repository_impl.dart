import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/cart_entity.dart';
import '../../domain/repositories/cart_repository.dart';
import '../datasources/cart_remote_datasource.dart';

class CartRepositoryImpl implements CartRepository {
  const CartRepositoryImpl(this._remote);
  final CartRemoteDataSource _remote;

  @override
  Future<Result<Cart>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<Cart>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Cart>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Cart>> syncToCloud({required Map<String, dynamic> body}) =>
      _remote.syncToCloud(body: body);

  @override
  Future<Result<Cart>> mergeAnonymousCart({required String anonymousCartId}) =>
      _remote.mergeAnonymousCart(anonymousCartId: anonymousCartId);

  @override
  Future<Result<Cart>> validateInventory() => _remote.validateInventory();

  @override
  Future<Result<Cart>> applyCoupon(String code) => _remote.applyCoupon(code);

  @override
  Future<Result<Cart>> removeCoupon() => _remote.removeCoupon();

  @override
  Future<Result<Cart>> applyGiftCard(String code) => _remote.applyGiftCard(code);

  @override
  Future<Result<Cart>> applyWallet(int amountMinor) => _remote.applyWallet(amountMinor);

  @override
  Future<Result<Cart>> applyLoyaltyPoints(int points) =>
      _remote.applyLoyaltyPoints(points);

  @override
  Future<Result<Cart>> estimate() => _remote.estimate();
}
