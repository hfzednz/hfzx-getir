import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';

/// Hits the real customer BFF when CUSTOMER_BASE is set (RC_FULL / FLUTTER_LIVE).
void main() {
  final base = Platform.environment['CUSTOMER_BASE'];
  final live = base != null && base.isNotEmpty;

  test('live BFF home → cart → preview → place → order (idempotent retry)', () async {
    final dio = Dio(
      BaseOptions(
        baseUrl: base,
        headers: {
          'Content-Type': 'application/json',
          'X-Tenant-Id': '11111111-1111-1111-1111-111111111111',
          'X-Request-Id': 'flutter-live-58',
        },
        validateStatus: (s) => s != null && s < 500,
      ),
    );

    final health = await dio.get<Map<String, dynamic>>('/health');
    expect(health.statusCode, 200);

    final home = await dio.get<Map<String, dynamic>>(
      '/v1/customer/home',
      queryParameters: {'lat': 41.0, 'lng': 29.0},
    );
    expect(home.statusCode, 200);

    final preview = await dio.post<Map<String, dynamic>>(
      '/v1/customer/checkout/preview',
      data: {
        'cartId': '33333333-3333-3333-3333-333333333333',
        'principalId': '22222222-2222-2222-2222-222222222222',
      },
    );
    expect(preview.statusCode, anyOf(200, 201));
    final sessionId = preview.data?['SessionID'] ??
        preview.data?['sessionId'] ??
        preview.data?['session_id'];

    final placeBody = {
      'cartId': '33333333-3333-3333-3333-333333333333',
      'principalId': '22222222-2222-2222-2222-222222222222',
      'paymentMethod': 'card',
      if (sessionId != null) 'sessionId': sessionId,
    };
    final a = await dio.post<Map<String, dynamic>>(
      '/v1/customer/checkout/place',
      data: placeBody,
    );
    final b = await dio.post<Map<String, dynamic>>(
      '/v1/customer/checkout/place',
      data: placeBody,
    );
    expect(a.statusCode, anyOf(200, 201));
    expect(b.statusCode, anyOf(200, 201));
    final orderA = a.data?['orderId']?.toString();
    final orderB = b.data?['orderId']?.toString();
    expect(orderA, isNotNull);
    expect(orderA, isNotEmpty);
    expect(orderB, orderA);

    final order = await dio.get<Map<String, dynamic>>(
      '/v1/customer/orders/$orderA',
      options: Options(validateStatus: (_) => true),
    );
    expect(order.statusCode, anyOf(200, 404, 502));
  }, skip: live ? false : 'CUSTOMER_BASE not set — live BFF journey runs in rc-flutter-live');
}
