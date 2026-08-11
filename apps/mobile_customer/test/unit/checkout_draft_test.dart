import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/features/addresses/domain/entities/addresses_entity.dart';
import 'package:nexora_customer/features/checkout/domain/entities/checkout_entity.dart';
import 'package:nexora_customer/shared/business_rules/checkout_rules.dart';

void main() {
  group('CheckoutDraft', () {
    test('defaults keep gift off and substitution allow', () {
      const draft = CheckoutDraft();
      expect(draft.gift, isFalse);
      expect(draft.giftMessage, isEmpty);
      expect(draft.substitutionPreference, SubstitutionPreference.allow);
      expect(draft.outOfStockRule, OutOfStockReplacementRule.similar);
      expect(draft.scheduleMode, CheckoutScheduleMode.asap);
    });

    test('props include gift and substitution fields', () {
      final draft = CheckoutDraft(
        address: const Address(id: 'a1', title: 'Home'),
        gift: true,
        giftMessage: 'Congrats',
        substitutionPreference: SubstitutionPreference.reject,
        outOfStockRule: OutOfStockReplacementRule.refund,
      );
      expect(draft.props, containsAll([true, 'Congrats', SubstitutionPreference.reject]));
    });
  });
}
