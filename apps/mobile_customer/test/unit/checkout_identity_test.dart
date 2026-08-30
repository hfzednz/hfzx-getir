import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Checkout is keyed by the cart id and the principal id. Both must come from the runtime
/// cart and the signed-in session; a constant would quote a cart the customer never filled
/// and, in the shared in-memory staging stack, one already consumed by someone else.
void main() {
  test('no production source hardcodes the seeded cart or principal', () {
    const seeded = [
      '33333333-3333-3333-3333-333333333333',
      '22222222-2222-2222-2222-222222222222',
    ];
    final offenders = <String>[];
    for (final entity in Directory('lib').listSync(recursive: true)) {
      if (entity is! File || !entity.path.endsWith('.dart')) continue;
      final source = entity.readAsStringSync();
      for (final id in seeded) {
        if (source.contains(id)) {
          offenders.add('${entity.path} contains $id');
        }
      }
    }
    expect(offenders, isEmpty);
  });

  test('checkout datasource has no identity fallback', () {
    final source = File(
      'lib/features/checkout/data/datasources/checkout_remote_datasource.dart',
    ).readAsStringSync();
    expect(source.contains('CheckoutBffDefaults'), isFalse);
    // A missing cart or session must surface as a failure rather than a substituted id.
    expect(source.contains('_missingIdentity'), isTrue);
  });
}
