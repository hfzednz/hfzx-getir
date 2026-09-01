import 'package:nexora_core/nexora_core.dart';

/// Password strength and policy validation.
abstract final class PasswordValidator {
  static const minLength = 8;
  static const maxLength = 128;

  static final _uppercase = RegExp(r'[A-Z]');
  static final _lowercase = RegExp(r'[a-z]');
  static final _digit = RegExp(r'\d');

  static String? validate(String? value, {int minLen = minLength}) {
    final result = parse(value, minLen: minLen);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message ?? 'Password does not meet requirements';
  }

  static Result<String> parse(String? raw, {int minLen = minLength}) {
    final password = raw ?? '';
    if (password.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password is required',
          details: {'field': 'password'},
        ),
      );
    }

    if (password.length < minLen) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password must be at least $minLen characters',
          details: {'field': 'password', 'min_length': minLen},
        ),
      );
    }

    if (password.length > maxLength) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password is too long',
          details: {'field': 'password', 'max_length': maxLength},
        ),
      );
    }

    if (!_uppercase.hasMatch(password)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password must include an uppercase letter',
          details: {'field': 'password', 'rule': 'uppercase'},
        ),
      );
    }

    if (!_lowercase.hasMatch(password)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password must include a lowercase letter',
          details: {'field': 'password', 'rule': 'lowercase'},
        ),
      );
    }

    if (!_digit.hasMatch(password)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Password must include a number',
          details: {'field': 'password', 'rule': 'digit'},
        ),
      );
    }

    return Success(password);
  }

  /// Returns 0–4 strength score for UI hints (not a security gate).
  static int strengthScore(String password) {
    var score = 0;
    if (password.length >= minLength) score++;
    if (password.length >= 12) score++;
    if (_uppercase.hasMatch(password) && _lowercase.hasMatch(password)) score++;
    if (_digit.hasMatch(password)) score++;
    return score.clamp(0, 4);
  }
}
