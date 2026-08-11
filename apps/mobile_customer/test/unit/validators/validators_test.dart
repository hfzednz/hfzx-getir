import 'package:flutter_test/flutter_test.dart';
import 'package:nexora_customer/features/addresses/domain/entities/addresses_entity.dart';
import 'package:nexora_customer/shared/validators/address_validator.dart';
import 'package:nexora_customer/shared/validators/age_restriction_validator.dart';
import 'package:nexora_customer/shared/validators/coupon_validator.dart';
import 'package:nexora_customer/shared/validators/email_validator.dart';
import 'package:nexora_customer/shared/validators/otp_validator.dart';
import 'package:nexora_customer/shared/validators/password_validator.dart';
import 'package:nexora_customer/shared/validators/phone_validator.dart';

void main() {
  group('PhoneValidator', () {
    test('normalizes TR mobile to E.164 digits', () {
      final result = PhoneValidator.parse('555 123 45 67');
      expect(result.isSuccess, isTrue);
      expect(result.valueOrNull, '905551234567');
    });

    test('rejects too-short numbers', () {
      expect(PhoneValidator.parse('123').isFailure, isTrue);
    });
  });

  group('EmailValidator', () {
    test('normalizes valid email', () {
      final result = EmailValidator.parse('User@Example.com');
      expect(result.isSuccess, isTrue);
      expect(result.valueOrNull, 'user@example.com');
    });

    test('rejects invalid email', () {
      expect(EmailValidator.parse('not-an-email').isFailure, isTrue);
    });
  });

  group('OtpValidator', () {
    test('accepts 6-digit code', () {
      expect(OtpValidator.parse('123456').isSuccess, isTrue);
    });

    test('rejects non-numeric code', () {
      expect(OtpValidator.parse('12ab56').isFailure, isTrue);
    });
  });

  group('PasswordValidator', () {
    test('accepts strong password', () {
      expect(PasswordValidator.parse('Secret1x').isSuccess, isTrue);
    });

    test('requires uppercase', () {
      expect(PasswordValidator.parse('secret1x').isFailure, isTrue);
    });
  });

  group('CouponValidator', () {
    test('uppercases valid code', () {
      final result = CouponValidator.parse('save-10');
      expect(result.isSuccess, isTrue);
      expect(result.valueOrNull, 'SAVE-10');
    });

    test('rejects empty code', () {
      expect(CouponValidator.parse('').isFailure, isTrue);
    });
  });

  group('AddressValidator', () {
    test('validates checkout address with serviceability', () {
      const address = Address(
        id: 'a1',
        title: 'Home',
        formatted: '123 Main Street, Istanbul',
        lat: 41.0,
        lng: 29.0,
      );
      final result = AddressValidator.validateForCheckout(address, serviceable: true);
      expect(result.isSuccess, isTrue);
    });

    test('rejects unserviceable address', () {
      const address = Address(
        id: 'a1',
        title: 'Home',
        formatted: '123 Main Street',
        lat: 41.0,
        lng: 29.0,
      );
      final result = AddressValidator.validateForCheckout(address, serviceable: false);
      expect(result.isFailure, isTrue);
    });
  });

  group('AgeRestrictionValidator', () {
    test('passes age gate when verified', () {
      final result = AgeRestrictionValidator.verifyAgeGate(
        cartHasAgeRestrictedItems: true,
        userAgeVerified: true,
      );
      expect(result.isSuccess, isTrue);
    });

    test('validates birth date for restricted cart', () {
      final birthDate = DateTime(1990, 1, 1);
      final result = AgeRestrictionValidator.verifyAgeGate(
        cartHasAgeRestrictedItems: true,
        userAgeVerified: false,
        birthDate: birthDate,
        referenceDate: DateTime(2026, 1, 1),
      );
      expect(result.isSuccess, isTrue);
    });

    test('rejects underage birth date', () {
      final birthDate = DateTime(2015, 1, 1);
      final result = AgeRestrictionValidator.verifyAgeGate(
        cartHasAgeRestrictedItems: true,
        userAgeVerified: false,
        birthDate: birthDate,
        referenceDate: DateTime(2026, 1, 1),
      );
      expect(result.isFailure, isTrue);
    });
  });
}
