/// Money formatting — minor units per CONSTITUTION §30.
class Money {
  const Money({required this.minorUnits, required this.currency});

  final int minorUnits;
  final String currency;

  double get majorUnits => minorUnits / 100;

  factory Money.fromJson(Map<String, dynamic> json) => Money(
        minorUnits: (json['minor_units'] as num?)?.toInt() ??
            (json['amount'] as num?)?.toInt() ??
            0,
        currency: json['currency']?.toString() ?? 'TRY',
      );

  Map<String, dynamic> toJson() => {
        'minor_units': minorUnits,
        'currency': currency,
      };

  String format({String? locale}) {
    final symbol = switch (currency.toUpperCase()) {
      'TRY' => '₺',
      'USD' => '\$',
      'EUR' => '€',
      _ => currency,
    };
    return '$symbol${majorUnits.toStringAsFixed(2)}';
  }
}
