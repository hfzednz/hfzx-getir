import 'package:nexora_core/nexora_core.dart';

/// Email validation and normalization.
abstract final class EmailValidator {
  static final _emailRegex = RegExp(r'^[\w.+-]+@[\w.-]+\.[a-zA-Z]{2,}$');

  static const maxLength = 254;

  static String? validate(String? value) {
    final result = parse(value);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message ?? 'Enter a valid email address';
  }

  static Result<String> parse(String? raw) {
    final email = raw?.trim() ?? '';
    if (email.isEmpty) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Email is required',
          details: {'field': 'email'},
        ),
      );
    }

    if (email.length > maxLength) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Email is too long',
          details: {'field': 'email', 'max_length': maxLength},
        ),
      );
    }

    final normalized = email.toLowerCase();
    if (!_emailRegex.hasMatch(normalized)) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Enter a valid email address',
          details: {'field': 'email'},
        ),
      );
    }

    return Success(normalized);
  }
}
