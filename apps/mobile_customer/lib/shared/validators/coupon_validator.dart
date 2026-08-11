import 'package:nexora_core/nexora_core.dart';

/// Promo / coupon code format validation (server applies eligibility rules).
abstract final class CouponValidator {
  static const minLength = 3;
  static const maxLength = 32;

  static final _codeRegex = RegExp(r'^[A-Z0-9][A-Z0-9\-_]{2,31}$');

  static String? validate(String? value) {
    final result = parse(value);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message ?? 'Enter a valid coupon code';
  }

  static Result<String> parse(String? raw) {
    final code = raw?.trim().toUpperCase() ?? '';
    if (code.isEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Coupon code is required',
          details: {'field': 'coupon_code'},
        ),
      );
    }

    if (code.length < minLength || code.length > maxLength) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Coupon code must be $minLength–$maxLength characters',
          details: {'field': 'coupon_code'},
        ),
      );
    }

    if (!_codeRegex.hasMatch(code)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Coupon code contains invalid characters',
          details: {'field': 'coupon_code'},
        ),
      );
    }

    return Success(code);
  }
}
