import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_courier/features/status/domain/entities/duty_status.dart';
import 'package:nexora_courier/shared/business_rules/duty_rules.dart';

void main() {
  group('DutyRules.canTransition', () {
    test('allows offline → online', () {
      expect(
        DutyRules.canTransition(DutyStatus.offline, DutyStatus.online),
        isTrue,
      );
    });

    test('blocks offline → busy', () {
      expect(
        DutyRules.canTransition(DutyStatus.offline, DutyStatus.busy),
        isFalse,
      );
    });

    test('allows online → onBreak', () {
      expect(
        DutyRules.canTransition(DutyStatus.online, DutyStatus.onBreak),
        isTrue,
      );
    });
  });

  group('DutyRules.validateTransition', () {
    test('succeeds for legal transition', () {
      final result = DutyRules.validateTransition(
        from: DutyStatus.online,
        to: DutyStatus.busy,
      );
      expect(result.isSuccess, isTrue);
    });

    test('fails when going offline with active delivery', () {
      final result = DutyRules.validateTransition(
        from: DutyStatus.busy,
        to: DutyStatus.offline,
        hasActiveDelivery: true,
      );
      expect(result.isFailure, isTrue);
    });

    test('fails illegal emergency → busy', () {
      final result = DutyRules.validateTransition(
        from: DutyStatus.emergency,
        to: DutyStatus.busy,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('DutyRules.canAcceptOffers', () {
    test('online and busy are eligible', () {
      expect(DutyRules.canAcceptOffers(DutyStatus.online), isTrue);
      expect(DutyRules.canAcceptOffers(DutyStatus.busy), isTrue);
      expect(DutyRules.canAcceptOffers(DutyStatus.offline), isFalse);
    });
  });
}
