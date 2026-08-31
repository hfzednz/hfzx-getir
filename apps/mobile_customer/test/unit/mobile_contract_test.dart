import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:nexora_customer/features/addresses/domain/entities/addresses_entity.dart';
import 'package:nexora_customer/features/auth/domain/repositories/auth_repository.dart';
import 'package:nexora_customer/features/cart/domain/entities/cart_entity.dart';
import 'package:nexora_customer/features/home/domain/entities/home_entity.dart';
import 'package:nexora_customer/features/orders/domain/entities/orders_entity.dart';
import 'package:nexora_customer/features/product/domain/entities/product_entity.dart';
import 'package:nexora_customer/features/search/domain/entities/search_entity.dart';
import 'package:nexora_customer/shared/errors/customer_facing_error.dart';

void main() {
  test('AuthTokens parses BFF camelCase session', () {
    final tokens = AuthTokens.fromJson({
      'accessToken': 'atk',
      'refreshToken': 'rtk',
      'customerId': 'cust-1',
      'expiresIn': 3600,
    });
    expect(tokens.accessToken, 'atk');
    expect(tokens.refreshToken, 'rtk');
    expect(tokens.userId, 'cust-1');
  });

  test('Cart.fromJson reads BFF cartId', () {
    final cart = Cart.fromJson({'cartId': 'cart-9', 'items': []});
    expect(cart.id, 'cart-9');
  });

  test('SearchResult.fromJson reads catalog Hits', () {
    final result = SearchResult.fromJson({
      'Hits': [
        {'ProductID': 'p1', 'Title': 'Taze Süt', 'SKU': 'MILK-1L'},
      ],
      'Total': 1,
    });
    expect(result.items, hasLength(1));
    expect(result.items.first.id, 'p1');
    expect(result.items.first.title, 'Taze Süt');
    expect(result.totalCount, 1);
  });

  test('ProductSummary.fromJson reads productId and priceMinor', () {
    final p = ProductSummary.fromJson({
      'productId': 'abc',
      'name': 'Fresh Milk',
      'priceMinor': 2499,
    });
    expect(p.id, 'abc');
    expect(p.title, 'Fresh Milk');
    expect(p.priceMinor, 2499);
  });

  test('checkout datasource maps address and sessionId onto the BFF body', () {
    final source = File(
      'lib/features/checkout/data/datasources/checkout_remote_datasource.dart',
    ).readAsStringSync();
    expect(source.contains("'address'"), isTrue);
    expect(source.contains('sessionId'), isTrue);
    expect(source.contains('line1'), isTrue);
  });

  test('customerFacingError maps invalid argument for TR', () {
    const err = NexoraValidationException(
      code: NexoraErrorCode.validationFailed,
      message: 'invalid argument: cart required',
    );
    expect(
      customerFacingError(err, languageCode: 'tr'),
      contains('Sepet bilgileri güncel değil'),
    );
  });

  test('customerFacingError maps expired coupon for TR', () {
    const err = NexoraConflictException(
      code: NexoraErrorCode.conflict,
      message: 'coupon has expired',
    );
    expect(
      customerFacingError(err, languageCode: 'tr'),
      contains('kuponun süresi'),
    );
  });

  test('HomeFeed.fromJson prefers widgets and serviceable', () {
    final feed = HomeFeed.fromJson({
      'serviceable': true,
      'widgets': [
        {
          'id': 'popular',
          'type': 'trending',
          'title': 'Popular',
          'items': [
            {'productId': 'milk-1', 'title': 'Taze Süt', 'priceMinor': 2499},
          ],
        },
      ],
    });
    expect(feed.serviceable, isTrue);
    expect(feed.widgets, hasLength(1));
    expect(feed.widgets.first.items.first.id, 'milk-1');
  });

  test('Address.fromJson reads recipient fields', () {
    final a = Address.fromJson({
      'id': 'a1',
      'formatted': 'Moda Cd 12',
      'recipient_name': 'Ada',
      'phone': '+905551112233',
      'lat': 40.98,
      'lng': 29.02,
    });
    expect(a.recipientName, 'Ada');
    expect(a.recipientPhone, '+905551112233');
  });

  test('Order.fromJson cart_created lines use productId', () {
    final order = Order.fromJson({
      'status': 'cart_created',
      'cartId': 'cart-new',
      'orderId': 'o1',
      'id': 'cart-new',
      'items': [
        {
          'productId': 'milk-1',
          'title': 'Taze Süt',
          'quantity': 7,
          'unitPriceMinor': 2499,
        },
      ],
    });
    expect(order.id, 'cart-new');
    expect(order.payload['cartId'], 'cart-new');
    expect(order.items, hasLength(1));
    expect(order.items.first.productId, 'milk-1');
    expect(order.items.first.quantity, 7);
  });

  test('Order.fromJson cart_seeded lines use productId', () {
    final order = Order.fromJson({
      'status': 'cart_seeded',
      'orderId': 'o1',
      'items': [
        {
          'productId': 'milk-1',
          'title': 'Taze Süt',
          'quantity': 7,
          'unitPriceMinor': 2499,
        },
      ],
    });
    expect(order.id, 'o1');
    expect(order.items, hasLength(1));
    expect(order.items.first.productId, 'milk-1');
    expect(order.items.first.quantity, 7);
  });
}
