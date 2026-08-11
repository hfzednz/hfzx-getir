import 'dart:convert';

import 'package:drift/drift.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../../shared/business_rules/cart_rules.dart';
import '../../data/datasources/cart_remote_datasource.dart';
import '../../data/local/app_database.dart';
import '../../data/repositories/cart_repository_impl.dart';
import '../../domain/entities/cart_entity.dart';
import '../../domain/repositories/cart_repository.dart';

final cartRemoteDataSourceProvider = Provider<CartRemoteDataSource>((ref) {
  return CartRemoteDataSource(ref.watch(apiClientProvider));
});

final cartCloudRepositoryProvider = Provider<CartRepository>((ref) {
  return CartRepositoryImpl(ref.watch(cartRemoteDataSourceProvider));
});

final cartItemsStreamProvider = StreamProvider<List<CartItem>>((ref) {
  return ref.watch(databaseProvider).watchCartItems();
});

final cartItemCountProvider = Provider<int>((ref) {
  return ref.watch(cartItemsStreamProvider).maybeWhen(
        data: (items) => items.fold(0, (sum, i) => sum + i.quantity),
        orElse: () => 0,
      );
});

final cartEstimateProvider = FutureProvider.autoDispose<Cart>((ref) async {
  final result = await ref.watch(cartCloudRepositoryProvider).estimate();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final cartLocalProvider = FutureProvider.autoDispose<Cart>((ref) async {
  return ref.watch(cartRepositoryProvider).buildLocalCart();
});

final cartRepositoryProvider = Provider<CartLocalRepository>((ref) {
  return CartLocalRepository(
    ref.watch(databaseProvider),
    ref.watch(apiClientProvider),
    ref.watch(cartCloudRepositoryProvider),
    ref.watch(mutationOutboxProvider),
    ref.watch(syncEngineProvider),
    ref.watch(analyticsTrackerProvider),
  );
});

/// Offline-first cart: Drift is local SoR; mutations enqueue to Hive outbox.
class CartLocalRepository {
  CartLocalRepository(
    this._db,
    this._client,
    this._remote,
    this._outbox,
    this._syncEngine,
    this._analytics,
  );

  final AppDatabase _db;
  final ApiClient _client;
  final CartRepository _remote;
  final MutationOutbox _outbox;
  final SyncEngine _syncEngine;
  final AnalyticsTracker _analytics;
  final _uuid = const Uuid();

  Future<void> _enqueueSync() async {
    final items = await _db.select(_db.cartItems).get();
    final body = {
      'items': items
          .map(
            (i) => {
              'product_id': i.productId,
              if (i.variantId != null) 'variant_id': i.variantId,
              'quantity': i.quantity,
              'unit_price_minor': i.unitPriceMinor,
              'currency': i.currency,
              if (i.notes != null) 'notes': i.notes,
            },
          )
          .toList(),
    };

    await _outbox.enqueue(
      PendingMutation(
        clientMutationId: _uuid.v4(),
        idempotencyKey: _uuid.v4(),
        method: 'POST',
        path: '/cart/sync',
        body: body,
        createdAt: DateTime.now().toUtc(),
      ),
    );

    // Best-effort immediate flush when online.
    // ignore: unawaited_futures
    _syncEngine.flush();
  }

  Future<void> addItem({
    required String productId,
    String? variantId,
    required String title,
    String? imageUrl,
    required int unitPriceMinor,
    String currency = 'TRY',
    int quantity = 1,
  }) async {
    await _db.upsertCartItem(
      CartItemsCompanion.insert(
        productId: productId,
        variantId: Value(variantId),
        title: title,
        imageUrl: Value(imageUrl),
        quantity: Value(quantity),
        unitPriceMinor: unitPriceMinor,
        currency: Value(currency),
        pendingSync: const Value(true),
      ),
    );
    // ignore: unawaited_futures
    _analytics.trackCartItemAdded(
      productId: productId,
      quantity: quantity,
      unitPriceMinor: unitPriceMinor,
      currency: currency,
    );
    await _enqueueSync();
  }

  Future<void> updateQuantity(
    String productId,
    int quantity, {
    String? variantId,
  }) async {
    if (quantity <= 0) {
      await _db.removeCartItem(productId, variantId: variantId);
      await _enqueueSync();
      return;
    }
    final items = await _db.select(_db.cartItems).get();
    final existing = items.firstWhere(
      (e) => e.productId == productId && e.variantId == variantId,
    );
    await _db.upsertCartItem(
      CartItemsCompanion(
        productId: Value(productId),
        variantId: Value(variantId),
        title: Value(existing.title),
        imageUrl: Value(existing.imageUrl),
        quantity: Value(quantity),
        unitPriceMinor: Value(existing.unitPriceMinor),
        currency: Value(existing.currency),
        pendingSync: const Value(true),
      ),
    );
    await _enqueueSync();
  }

  Future<Cart> buildLocalCart() async {
    final items = await _db.select(_db.cartItems).get();
    final lines = items
        .map(
          (i) => CartLine(
            productId: i.productId,
            variantId: i.variantId,
            title: i.title,
            imageUrl: i.imageUrl,
            quantity: i.quantity,
            unitPriceMinor: i.unitPriceMinor,
            currency: i.currency,
          ),
        )
        .toList();
    final subtotal = lines.fold<int>(0, (sum, l) => sum + l.lineTotalMinor);
    return Cart(
      id: 'local',
      items: lines,
      subtotalMinor: subtotal,
      totalMinor: subtotal,
      currency: items.isEmpty ? 'TRY' : items.first.currency,
    );
  }

  Future<Result<Cart>> syncToCloud() async {
    final items = await _db.select(_db.cartItems).get();
    final body = {
      'items': items
          .map(
            (i) => {
              'product_id': i.productId,
              if (i.variantId != null) 'variant_id': i.variantId,
              'quantity': i.quantity,
              'unit_price_minor': i.unitPriceMinor,
              'currency': i.currency,
            },
          )
          .toList(),
    };
    final result = await _remote.syncToCloud(body: body);
    if (result case Success(:final value)) {
      for (final line in value.items) {
        await _db.upsertCartItem(
          CartItemsCompanion.insert(
            productId: line.productId,
            variantId: Value(line.variantId),
            title: line.title,
            imageUrl: Value(line.imageUrl),
            quantity: Value(line.quantity),
            unitPriceMinor: line.unitPriceMinor,
            currency: Value(line.currency),
            pendingSync: const Value(false),
          ),
        );
      }
      return Success(value);
    }
    return result;
  }

  Future<Result<Cart>> mergeAnonymousCart(String anonymousCartId) =>
      _remote.mergeAnonymousCart(anonymousCartId: anonymousCartId);

  Future<Result<Cart>> validateInventory() => _remote.validateInventory();

  Future<Result<Cart>> applyCoupon(String code) => _remote.applyCoupon(code);

  Future<Result<Cart>> removeCoupon() => _remote.removeCoupon();

  Future<Result<Cart>> applyGiftCard(String code) =>
      _remote.applyGiftCard(code);

  Future<Result<Cart>> applyWallet(int amountMinor) =>
      _remote.applyWallet(amountMinor);

  Future<Result<Cart>> applyLoyaltyPoints(int points) =>
      _remote.applyLoyaltyPoints(points);

  Future<Result<Cart>> estimate() => _remote.estimate();

  Future<void> validateWithServer() async {
    await _client.post<void>('/cart/validate');
  }

  /// Debug helper — serializes pending Drift rows.
  Future<String> pendingSyncDebugJson() async {
    final pending = await (_db.select(_db.cartItems)
          ..where((t) => t.pendingSync.equals(true)))
        .get();
    return jsonEncode(
      pending
          .map(
            (i) => {
              'product_id': i.productId,
              'quantity': i.quantity,
            },
          )
          .toList(),
    );
  }

  Cart withRules(Cart cart) {
    final violations = [...cart.violations, ...CartRules.evaluate(cart)];
    return Cart(
      id: cart.id,
      items: cart.items,
      promotions: cart.promotions,
      coupon: cart.coupon,
      giftCards: cart.giftCards,
      walletAppliedMinor: cart.walletAppliedMinor,
      loyaltyPointsToRedeem: cart.loyaltyPointsToRedeem,
      subtotalMinor: cart.subtotalMinor,
      deliveryFeeEstimateMinor: cart.deliveryFeeEstimateMinor,
      taxEstimateMinor: cart.taxEstimateMinor,
      totalMinor: cart.totalMinor,
      currency: cart.currency,
      etaMinutes: cart.etaMinutes,
      minOrderMinor: cart.minOrderMinor,
      violations: violations,
    );
  }
}

const anonymousCartIdPrefsKey = 'anonymous_cart_id';

/// Merges guest/anonymous cart into the authenticated cart after login.
Future<void> mergeCartAfterLogin(Ref ref) async {
  final prefs = ref.read(preferencesStoreProvider);
  final anonymousId = prefs.get<String>(anonymousCartIdPrefsKey);
  if (anonymousId == null || anonymousId.isEmpty) return;

  final cart = ref.read(cartRepositoryProvider);
  await cart.mergeAnonymousCart(anonymousId);
  await cart.syncToCloud();
  await prefs.remove(anonymousCartIdPrefsKey);
}
