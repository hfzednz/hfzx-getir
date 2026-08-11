import 'package:uuid/uuid.dart';

/// Generates idempotency keys for money/inventory mutations (CONSTITUTION §30).
class Idempotency {
  Idempotency._();

  static const _uuid = Uuid();

  static String generate() => _uuid.v4();
}
