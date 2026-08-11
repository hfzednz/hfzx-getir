import 'package:flutter/material.dart';
import 'package:nexora_design/nexora_design.dart';

/// Minimal placeholder used by CORE routes until feature screens land.
class PlaceholderScreen extends StatelessWidget {
  const PlaceholderScreen({
    super.key,
    required this.title,
    this.subtitle,
  });

  final String title;
  final String? subtitle;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: NxTopBar(title: title),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(title, style: NxTypography.titleLg),
              if (subtitle != null) ...[
                const SizedBox(height: NxSpacing.s2),
                Text(
                  subtitle!,
                  textAlign: TextAlign.center,
                  style: NxTypography.bodyMd,
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
