import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/bootstrap/deep_link_handler.dart';

void main() {
  group('DeepLinkHandler.mapUriToLocation', () {
    test('maps https product path', () {
      final uri = Uri.parse('https://nexora.app/p/sku-123');
      expect(DeepLinkHandler.mapUriToLocation(uri), '/p/sku-123');
    });

    test('maps nexora scheme product host', () {
      final uri = Uri.parse('nexora://product/sku-99');
      expect(DeepLinkHandler.mapUriToLocation(uri), '/p/sku-99');
    });

    test('maps order tracking', () {
      final uri = Uri.parse('https://nexora.app/orders/ord-1/track');
      expect(DeepLinkHandler.mapUriToLocation(uri), '/orders/ord-1/track');
    });

    test('maps promo', () {
      final uri = Uri.parse('https://nexora.app/promo/WELCOME10');
      expect(DeepLinkHandler.mapUriToLocation(uri), '/promo/WELCOME10');
    });

    test('returns null for unknown', () {
      final uri = Uri.parse('https://nexora.app/unknown/path');
      expect(DeepLinkHandler.mapUriToLocation(uri), isNull);
    });
  });
}
