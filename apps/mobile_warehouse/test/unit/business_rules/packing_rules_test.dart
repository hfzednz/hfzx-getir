import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_warehouse/features/packing/domain/entities/pack_task.dart';
import 'package:nexora_warehouse/shared/business_rules/packing_rules.dart';

void main() {
  group('PackingRules.validateWeight', () {
    test('accepts weight within tolerance', () {
      final result = PackingRules.validateWeight(
        actualGrams: 1020,
        expectedGrams: 1000,
      );
      expect(result.isSuccess, isTrue);
    });

    test('rejects weight outside tolerance', () {
      final result = PackingRules.validateWeight(
        actualGrams: 1200,
        expectedGrams: 1000,
      );
      expect(result.isFailure, isTrue);
    });

    test('rejects non-positive actual', () {
      final result = PackingRules.validateWeight(
        actualGrams: 0,
        expectedGrams: 1000,
      );
      expect(result.isFailure, isTrue);
    });
  });

  group('PackingRules.validateSeal', () {
    test('requires printed label', () {
      final result = PackingRules.validateSeal(
        status: PackTaskStatus.labeled,
        sealed: true,
        labelPrinted: false,
      );
      expect(result.isFailure, isTrue);
    });

    test('succeeds when labeled and sealed', () {
      final result = PackingRules.validateSeal(
        status: PackTaskStatus.labeled,
        sealed: true,
        labelPrinted: true,
      );
      expect(result.isSuccess, isTrue);
    });
  });

  group('PackingRules.validateTransition', () {
    test('allows packing → weighed', () {
      expect(
        PackingRules.validateTransition(
          from: PackTaskStatus.packing,
          to: PackTaskStatus.weighed,
        ).isSuccess,
        isTrue,
      );
    });

    test('blocks ready → packed', () {
      expect(
        PackingRules.validateTransition(
          from: PackTaskStatus.readyToPack,
          to: PackTaskStatus.packed,
        ).isFailure,
        isTrue,
      );
    });
  });
}
