import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:nexora_customer/features/addresses/domain/entities/addresses_entity.dart';
import 'package:nexora_customer/features/checkout/domain/entities/checkout_entity.dart';
import 'package:nexora_customer/shared/business_rules/checkout_rules.dart';

Address _address({String id = 'addr-1', bool serviceable = true}) => Address(
      id: id,
      title: 'Home',
      formatted: 'Test St 1',
      lat: 41.01,
      lng: 28.97,
      serviceable: serviceable,
    );

CheckoutDraft _draft({
  Address? address,
  bool gift = false,
  String giftMessage = '',
  SubstitutionPreference substitutionPreference = SubstitutionPreference.allow,
  OutOfStockReplacementRule outOfStockRule = OutOfStockReplacementRule.similar,
}) =>
    CheckoutDraft(
      address: address ?? _address(),
      gift: gift,
      giftMessage: giftMessage,
      substitutionPreference: substitutionPreference,
      outOfStockRule: outOfStockRule,
    );

CheckoutQuote _quote({
  int totalMinor = 1000,
  String currency = 'TRY',
  DateTime? expiresAt,
}) =>
    CheckoutQuote(
      subtotalMinor: totalMinor,
      deliveryFeeMinor: 0,
      discountMinor: 0,
      taxMinor: 0,
      totalMinor: totalMinor,
      currency: currency,
      quoteId: 'q-1',
      expiresAt: expiresAt,
    );

void main() {
  group('CheckoutRules.validateDraft', () {
    test('fails when gift is true without gift message', () {
      final result = CheckoutRules.validateDraft(
        draft: _draft(gift: true, giftMessage: '   '),
        addressServiceable: true,
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull, isA<NexoraValidationException>());
      expect(
        (result.errorOrNull as NexoraValidationException).details?['field'],
        'gift_message',
      );
    });

    test('succeeds when gift message is provided', () {
      final result = CheckoutRules.validateDraft(
        draft: _draft(gift: true, giftMessage: 'Happy birthday'),
        addressServiceable: true,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails on substitution reject + similar out-of-stock conflict', () {
      final result = CheckoutRules.validateDraft(
        draft: _draft(
          substitutionPreference: SubstitutionPreference.reject,
          outOfStockRule: OutOfStockReplacementRule.similar,
        ),
        addressServiceable: true,
      );
      expect(result.isFailure, isTrue);
      expect(
        result.errorOrNull!.message,
        contains('Cannot auto-replace'),
      );
    });
  });

  group('CheckoutRules.validateSubstitutionPreferences', () {
    test('allows reject with refund replacement', () {
      final result = CheckoutRules.validateSubstitutionPreferences(
        substitutionPreference: SubstitutionPreference.reject,
        outOfStockRule: OutOfStockReplacementRule.refund,
      );
      expect(result.isSuccess, isTrue);
    });
  });

  group('CheckoutRules.verifyFinalPrice', () {
    test('succeeds within tolerance', () {
      final result = CheckoutRules.verifyFinalPrice(
        clientQuote: _quote(totalMinor: 1000),
        serverQuote: _quote(totalMinor: 1001),
      );
      expect(result.isSuccess, isTrue);
      expect(result.valueOrNull!.totalMinor, 1001);
    });

    test('fails when delta exceeds tolerance', () {
      final result = CheckoutRules.verifyFinalPrice(
        clientQuote: _quote(totalMinor: 1000),
        serverQuote: _quote(totalMinor: 1005),
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('Order total changed'));
    });

    test('fails on currency mismatch', () {
      final result = CheckoutRules.verifyFinalPrice(
        clientQuote: _quote(currency: 'TRY'),
        serverQuote: _quote(currency: 'USD'),
      );
      expect(result.isFailure, isTrue);
      expect(result.errorOrNull!.message, contains('currency'));
    });
  });
}
