import 'package:nexora_core/nexora_core.dart';

import '../entities/cart_entity.dart';

abstract class CartRepository {
  Future<Result<Cart>> fetch({String? id});
  Future<Result<List<Cart>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Cart>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<Cart>> syncToCloud({required Map<String, dynamic> body});
  Future<Result<Cart>> mergeAnonymousCart({required String anonymousCartId});
  Future<Result<Cart>> validateInventory();
  Future<Result<Cart>> applyCoupon(String code);
  Future<Result<Cart>> removeCoupon();
  Future<Result<Cart>> applyGiftCard(String code);
  Future<Result<Cart>> applyWallet(int amountMinor);
  Future<Result<Cart>> applyLoyaltyPoints(int points);
  Future<Result<Cart>> estimate();
}
