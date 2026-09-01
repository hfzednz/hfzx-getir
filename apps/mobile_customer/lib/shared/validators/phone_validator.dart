import 'package:nexora_core/nexora_core.dart';

/// Phone number validation and normalization (TR + international).
abstract final class PhoneValidator {
  static const defaultCountryCode = '90';

  static final _trMobileRegex = RegExp(r'^5\d{9}$');
  static final _digitsOnly = RegExp(r'\D');

  /// Form field validator — returns user-facing message or null if valid.
  static String? validate(String? value, {String defaultCountryCode = defaultCountryCode}) {
    final result = parse(value, defaultCountryCode: defaultCountryCode);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message ?? 'Enter a valid phone number';
  }

  /// Parses and normalizes to E.164 digits without '+' (e.g. `905551234567`).
  static Result<String> parse(
    String? raw, {
    String defaultCountryCode = defaultCountryCode,
  }) {
    final input = raw?.trim() ?? '';
    if (input.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Phone number is required',
          details: {'field': 'phone'},
        ),
      );
    }

    var digits = input.replaceAll(_digitsOnly, '');
    if (digits.startsWith('00')) {
      digits = digits.substring(2);
    }

    if (digits.length == 10 && _trMobileRegex.hasMatch(digits)) {
      digits = '$defaultCountryCode$digits';
    } else if (digits.length == 11 && digits.startsWith('0') && _trMobileRegex.hasMatch(digits.substring(1))) {
      digits = '$defaultCountryCode${digits.substring(1)}';
    }

    if (digits.length < 10 || digits.length > 15) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Enter a valid phone number',
          details: {'field': 'phone', 'digits': digits.length},
        ),
      );
    }

    if (digits.startsWith(defaultCountryCode)) {
      final national = digits.substring(defaultCountryCode.length);
      if (!_trMobileRegex.hasMatch(national)) {
        return Failure(
          NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Enter a valid mobile number',
            details: {'field': 'phone', 'country': defaultCountryCode},
          ),
        );
      }
    }

    return Success(digits);
  }

  /// Formats E.164 digits for display (TR: +90 555 123 45 67).
  static String formatDisplay(String e164Digits, {String countryCode = defaultCountryCode}) {
    if (e164Digits.startsWith(countryCode) && e164Digits.length == countryCode.length + 10) {
      final n = e164Digits.substring(countryCode.length);
      return '+$countryCode ${n.substring(0, 3)} ${n.substring(3, 6)} ${n.substring(6, 8)} ${n.substring(8)}';
    }
    return '+$e164Digits';
  }
}
