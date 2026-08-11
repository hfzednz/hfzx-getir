import 'package:flutter_test/flutter_test.dart';

import 'package:nexora_customer/features/tracking/domain/entities/tracking_entity.dart';

void main() {
  group('TrackingSnapshot', () {
    test('fromJson maps courier and coords', () {
      final snap = TrackingSnapshot.fromJson({
        'order_id': 'ord-1',
        'status': 'in_transit',
        'eta_minutes': 12,
        'eta_min': 10,
        'eta_max': 15,
        'courier_name': 'Ada',
        'courier_phone': '+905551112233',
        'courier_lat': 41.01,
        'courier_lng': 28.97,
        'store_lat': 41.02,
        'store_lng': 28.98,
        'dest_lat': 41.0,
        'dest_lng': 29.0,
        'can_call': true,
        'can_chat': true,
        'steps': [
          {
            'title': 'Picked up',
            'subtitle': 'Courier has your order',
            'state': 'completed',
          },
          {
            'title': 'On the way',
            'state': 'current',
          },
        ],
      });

      expect(snap.orderId, 'ord-1');
      expect(snap.etaMin, 10);
      expect(snap.etaMax, 15);
      expect(snap.courierName, 'Ada');
      expect(snap.canCall, isTrue);
      expect(snap.steps, hasLength(2));
      expect(snap.steps.first.state, 'completed');
      expect(snap.hasMapCoordinates, isTrue);
    });

    test('fromJson parses route points and courier_chat_url', () {
      final snap = TrackingSnapshot.fromJson({
        'order_id': 'ord-2',
        'status': 'in_transit',
        'courier_chat_url': 'https://chat.example/c/1',
        'route': [
          {'lat': 41.0, 'lng': 29.0},
          {'lat': 41.01, 'lng': 29.01},
        ],
      });

      expect(snap.routePoints, hasLength(2));
      expect(snap.routePoints.first.lat, 41.0);
      expect(snap.courierChatUrl, 'https://chat.example/c/1');
      expect(snap.hasMapCoordinates, isTrue);
    });
  });
}
