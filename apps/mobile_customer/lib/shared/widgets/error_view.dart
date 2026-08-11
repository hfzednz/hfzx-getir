import 'package:flutter/material.dart';
import 'package:nexora_design/nexora_design.dart';

class ErrorView extends StatelessWidget {
  const ErrorView({
    super.key,
    required this.message,
    this.onRetry,
    this.title = 'Something went wrong',
  });

  final String title;
  final String message;
  final VoidCallback? onRetry;

  @override
  Widget build(BuildContext context) {
    return NxEmptyState(
      title: title,
      body: message,
      primaryActionLabel: onRetry != null ? 'Retry' : null,
      onPrimaryAction: onRetry,
    );
  }
}
