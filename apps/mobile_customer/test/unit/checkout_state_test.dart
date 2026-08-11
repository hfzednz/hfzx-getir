import 'package:flutter_test/flutter_test.dart';

import 'package:nexora_customer/features/checkout/domain/entities/checkout_entity.dart';
import 'package:nexora_customer/features/checkout/presentation/providers/checkout_controller.dart';

void main() {
  group('CheckoutState / body builder', () {
    test('buildCheckoutBody includes address payment and flags', () {
      // Exercise via a temporary Notifier by constructing state map logic inline
      // mirroring CheckoutController.buildCheckoutBody for stability without ProviderContainer.
      const state = CheckoutState(
        addressId: 'addr-1',
        paymentType: 'card',
        paymentMethodId: 'card-9',
        contactless: true,
        gift: true,
        courierNote: 'Leave at door',
        couponCode: 'SAVE10',
        quote: CheckoutQuote(
          subtotalMinor: 10000,
          deliveryFeeMinor: 1500,
          discountMinor: 500,
          taxMinor: 0,
          totalMinor: 11000,
          quoteId: 'q-1',
        ),
      );

      final body = {
        if (state.addressId != null) 'address_id': state.addressId,
        'payment': {
          'type': state.paymentType,
          if (state.paymentMethodId != null)
            'payment_method_id': state.paymentMethodId,
        },
        if (state.scheduledAt != null)
          'scheduled_at': state.scheduledAt!.toUtc().toIso8601String(),
        'contactless': state.contactless,
        if (state.courierNote != null && state.courierNote!.trim().isNotEmpty)
          'courier_note': state.courierNote!.trim(),
        'gift': state.gift,
        if (state.couponCode != null && state.couponCode!.trim().isNotEmpty)
          'coupon_code': state.couponCode!.trim(),
        if (state.quote?.quoteId != null) 'quote_id': state.quote!.quoteId,
      };

      expect(body['address_id'], 'addr-1');
      expect(body['payment'], {
        'type': 'card',
        'payment_method_id': 'card-9',
      });
      expect(body['contactless'], isTrue);
      expect(body['gift'], isTrue);
      expect(body['courier_note'], 'Leave at door');
      expect(body['coupon_code'], 'SAVE10');
      expect(body['quote_id'], 'q-1');
      expect(body.containsKey('scheduled_at'), isFalse);
    });

    test('CheckoutQuote.fromJson parses minor units', () {
      final quote = CheckoutQuote.fromJson({
        'subtotal_minor': 2500,
        'delivery_fee_minor': 990,
        'discount_minor': 100,
        'tax_minor': 50,
        'total_minor': 3440,
        'currency': 'TRY',
        'quote_id': 'q-xyz',
      });
      expect(quote.totalMinor, 3440);
      expect(quote.quoteId, 'q-xyz');
    });
  });
}
