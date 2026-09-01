import 'package:nexora_core/nexora_core.dart';

/// One-time password / verification code validation.
abstract final class OtpValidator {
  static const defaultLength = 6;

  static String? validate(String? value, {int length = defaultLength}) {
    final result = parse(value, length: length);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message ?? 'Enter a valid verification code';
  }

  static Result<String> parse(String? raw, {int length = defaultLength}) {
    final code = raw?.trim().replaceAll(RegExp(r'\s'), '') ?? '';
    if (code.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Verification code is required',
          details: {'field': 'otp'},
        ),
      );
    }

    if (code.length != length || int.tryParse(code) == null) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Enter a valid $length-digit code',
          details: {'field': 'otp', 'expected_length': length},
        ),
      );
    }

    return Success(code);
  }
}
