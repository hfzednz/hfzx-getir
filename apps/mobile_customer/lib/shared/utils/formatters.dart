import 'package:intl/intl.dart';

import 'money.dart';

abstract final class Formatters {
  static String dateTime(DateTime dt, {String? locale}) {
    return DateFormat.yMMMd(locale).add_jm().format(dt.toLocal());
  }

  static String etaMinutes(int minutes, {String? locale}) {
    if (minutes <= 0) return '—';
    return '$minutes min';
  }

  static String money(Money money, {String? locale}) => money.format(locale: locale);

  static String phone(String raw) {
    final digits = raw.replaceAll(RegExp(r'\D'), '');
    if (digits.length == 10) {
      return '(${digits.substring(0, 3)}) ${digits.substring(3, 6)} ${digits.substring(6)}';
    }
    return raw;
  }
}
