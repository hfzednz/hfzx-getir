import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/features/home/domain/entities/home_entity.dart';
import 'package:nexora_customer/features/home/presentation/home_history.dart';
import 'package:nexora_customer/features/orders/domain/entities/orders_entity.dart';
import 'package:nexora_customer/features/product/domain/entities/product_entity.dart';
import 'package:nexora_customer/features/search/domain/entities/search_entity.dart';
import 'package:nexora_customer/features/stores/domain/entities/store_entity.dart';
import 'package:nexora_customer/features/tracking/domain/entities/tracking_entity.dart';
import 'package:nexora_customer/shared/realtime/sse_parser.dart';
import 'package:nexora_core/nexora_core.dart';

void main() {
  test('SSE parser splits frames and ignores comments', () {
    const chunk = ': ping\n\nevent: order\ndata: {"status":"picking"}\n\ndata: leftover';
    final parsed = parseSseBuffer(chunk);
    expect(parsed.frames, hasLength(1));
    expect(parsed.frames.first.event, 'order');
    expect(parsed.frames.first.data, contains('picking'));
    expect(parsed.rest, 'data: leftover');
  });

  test('store-scoped product JSON marks out of stock', () {
    final p = ProductSummary.fromJson({
      'id': 'milk-1',
      'title': 'Taze Süt',
      'priceMinor': 1999,
      'outOfStock': true,
    });
    expect(p.stockStatus, ProductStockStatus.outOfStock);
  });

  test('search filters send storeId', () {
    const filters = SearchFilters(storeId: 'store-kadikoy');
    expect(filters.toQueryParams()['storeId'], 'store-kadikoy');
  });

  test('store summary parses closed warehouse', () {
    final store = StoreSummary.fromJson({
      'id': 'store-bakirkoy',
      'name': 'Bakırköy',
      'status': 'closed',
      'open': false,
      'minOrderMinor': 10000,
      'deliveryFeeMinor': 1999,
    });
    expect(store.open, isFalse);
    expect(store.minOrderMinor, 10000);
  });

  test('history widgets derive from real order lines', () {
    final widgets = historyWidgetsFromOrders([
      Order.fromJson({
        'orderId': 'o1',
        'status': 'delivered',
        'items': [
          {'productId': 'milk-1', 'title': 'Taze Süt', 'quantity': 7, 'unitPriceMinor': 1999},
          {'productId': 'bread-1', 'title': 'Ekmek', 'quantity': 1, 'unitPriceMinor': 1299},
        ],
      }),
      Order.fromJson({
        'orderId': 'o2',
        'status': 'delivered',
        'items': [
          {'productId': 'milk-1', 'title': 'Taze Süt', 'quantity': 2, 'unitPriceMinor': 1999},
        ],
      }),
    ]);
    expect(widgets.first.id, 'recently-ordered');
    expect(widgets.first.items.map((e) => e.id), contains('milk-1'));
    expect(widgets.last.id, 'frequently-purchased');
  });

  test('tracking lifecycle ranks pick before pack', () {
    expect(trackingStatusRank('picking'), lessThan(trackingStatusRank('packing')));
    expect(normalizeTrackingStatus('PICK'), 'picking');
    expect(normalizeTrackingStatus('delivered'), 'completed');
    final steps = trackingLifecycleSteps('picking');
    expect(steps[2].state, 'current');
    expect(steps[1].state, 'completed');
    expect(steps[3].state, 'upcoming');
  });

  test('sse url is derived from websocket base', () {
    const env = AppEnvironment.staging();
    expect(env.sseUrl, 'https://realtime.staging.nexora.io/v1/realtime/sse');
  });

  test('realtime ticket parses BFF payload', () {
    final ticket = RealtimeTicket.fromJson(
      {'ticket': 'abc', 'topic': 'order:o1', 'expiresIn': 120},
      fallbackTopic: 'order:fallback',
    );
    expect(ticket.ticket, 'abc');
    expect(ticket.topic, 'order:o1');
  });
}
