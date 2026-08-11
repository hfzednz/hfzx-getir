import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../providers/auth_controller.dart';
import '../providers/auth_session_provider.dart';

/// Requires clock-in before entering the warehouse shell.
class ShiftGateScreen extends ConsumerStatefulWidget {
  const ShiftGateScreen({super.key});

  @override
  ConsumerState<ShiftGateScreen> createState() => _ShiftGateScreenState();
}

class _ShiftGateScreenState extends ConsumerState<ShiftGateScreen> {
  final _stationController = TextEditingController();
  String? _error;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final local = ref.read(warehouseLocalStoreProvider);
      final last = local.lastStationId;
      if (last != null && last.isNotEmpty) {
        _stationController.text = last;
      }
    });
  }

  @override
  void dispose() {
    _stationController.dispose();
    super.dispose();
  }

  Future<void> _clockIn() async {
    setState(() => _error = null);
    final stationId = _stationController.text.trim();
    final session = await ref.read(authControllerProvider.notifier).clockIn(
          stationId: stationId.isEmpty ? null : stationId,
        );
    if (!mounted) return;
    if (session == null) {
      setState(() => _error = ref.read(authControllerProvider).error ?? 'Clock-in failed');
      return;
    }
    if (stationId.isNotEmpty) {
      await ref.read(warehouseLocalStoreProvider).setLastStationId(stationId);
    }
    if (!mounted) return;
    context.go(RouteNames.home);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final auth = ref.watch(authControllerProvider);
    final session = ref.watch(authSessionProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.clockIn),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s4),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              session.displayName ?? session.phone ?? '',
              style: NxTypography.titleMd,
            ),
            const SizedBox(height: NxSpacing.s2),
            Text(
              'Store: ${session.storeId ?? '—'}',
              style: NxTypography.bodyMd,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxField(
              controller: _stationController,
              label: 'Station ID',
              error: _error,
            ),
            const SizedBox(height: NxSpacing.s4),
            NxButton(
              label: l10n.clockIn,
              expand: true,
              loading: auth.isLoading,
              onPressed: _clockIn,
            ),
          ],
        ),
      ),
    );
  }
}
