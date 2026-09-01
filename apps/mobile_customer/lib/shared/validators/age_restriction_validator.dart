import 'package:nexora_core/nexora_core.dart';

/// Age-gate validation for alcohol, tobacco, and restricted categories.
abstract final class AgeRestrictionValidator {
  static const minimumAgeYears = 18;

  static String? validateBirthDate(DateTime? birthDate, {DateTime? referenceDate}) {
    final result = parseBirthDate(birthDate, referenceDate: referenceDate);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message;
  }

  static Result<int> parseBirthDate(DateTime? birthDate, {DateTime? referenceDate}) {
    if (birthDate == null) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Date of birth is required for age-restricted items',
          details: {'field': 'birth_date'},
        ),
      );
    }

    final now = referenceDate ?? DateTime.now();
    if (birthDate.isAfter(now)) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Date of birth cannot be in the future',
          details: {'field': 'birth_date'},
        ),
      );
    }

    final age = _ageInYears(birthDate, now);
    if (age < minimumAgeYears) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'You must be at least $minimumAgeYears to purchase restricted items',
          details: {'field': 'birth_date', 'minimum_age': minimumAgeYears, 'age': age},
        ),
      );
    }

    return Success(age);
  }

  static Result<void> verifyAgeGate({
    required bool cartHasAgeRestrictedItems,
    required bool userAgeVerified,
    DateTime? birthDate,
    DateTime? referenceDate,
  }) {
    if (!cartHasAgeRestrictedItems) {
      return const Success(null);
    }

    if (userAgeVerified) {
      return const Success(null);
    }

    if (birthDate != null) {
      final ageResult = parseBirthDate(birthDate, referenceDate: referenceDate);
      if (ageResult.isSuccess) {
        return const Success(null);
      }
      return Failure(ageResult.errorOrNull!);
    }

    return const Failure(
      NexoraValidationException(
        code: NexoraErrorCode.validationFailed,
        message: 'Age verification is required for restricted items',
        details: {'field': 'age_verified', 'minimum_age': minimumAgeYears},
      ),
    );
  }

  static int _ageInYears(DateTime birthDate, DateTime reference) {
    var years = reference.year - birthDate.year;
    if (reference.month < birthDate.month ||
        (reference.month == birthDate.month && reference.day < birthDate.day)) {
      years--;
    }
    return years;
  }
}
