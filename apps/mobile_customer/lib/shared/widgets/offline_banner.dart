import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../di/providers.dart';

class OfflineBanner extends ConsumerWidget {
  const OfflineBanner({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final onlineAsync = ref.watch(connectivityOnlineProvider);
    final online = onlineAsync.value ?? true;
    if (online) return const SizedBox.shrink();

    final colors = context.nxColors;
    return Material(
      color: colors.warning.withValues(alpha: 0.15),
      child: SafeArea(
        bottom: false,
        child: Padding(
          padding: const EdgeInsets.symmetric(
            horizontal: NxSpacing.s4,
            vertical: NxSpacing.s2,
          ),
          child: Row(
            children: [
              Icon(Icons.cloud_off, color: colors.warning, size: 18),
              const SizedBox(width: NxSpacing.s2),
              Expanded(
                child: Text(
                  'You are offline. Some actions are unavailable.',
                  style: NxTypography.bodySm.copyWith(color: colors.textPrimary),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
