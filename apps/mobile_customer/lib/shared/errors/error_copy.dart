import 'package:flutter/widgets.dart';

import 'customer_facing_error.dart';

/// Locale-aware customer copy for thrown errors.
String localizedCustomerError(BuildContext context, Object err) {
  return customerFacingError(
    err,
    languageCode: Localizations.localeOf(context).languageCode,
  );
}
