import 'package:flutter/material.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../l10n/app_localizations.dart';

class ErrorView extends StatelessWidget {
  const ErrorView({
    super.key,
    required this.message,
    this.onRetry,
    this.title,
  });

  final String? title;
  final String message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return NxEmptyState(
      title: title ?? l10n.somethingWentWrong,
      body: message,
      primaryActionLabel: onRetry != null ? l10n.retry : null,
      onPrimaryAction: onRetry,
    );
  }
}
