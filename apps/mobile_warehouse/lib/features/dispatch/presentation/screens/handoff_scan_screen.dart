import 'package:collection/collection.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/handoff_task.dart';
import '../providers/dispatch_providers.dart';

class HandoffScanScreen extends ConsumerStatefulWidget {
  const HandoffScanScreen({super.key, required this.handoffId});
  final String handoffId;

  @override
  ConsumerState<HandoffScanScreen> createState() => _HandoffScanScreenState();
}

class _HandoffScanScreenState extends ConsumerState<HandoffScanScreen> {
  bool _handling = false;
  String? _error;

  Future<void> _onDetect(BarcodeCapture capture, HandoffTask task) async {
    if (_handling) return;
    final raw = capture.barcodes.firstOrNull?.rawValue;
    if (raw == null || raw.isEmpty) return;
    setState(() { _handling = true; _error = null; });
    final result = await ref.read(dispatchActionsProvider).scan(
          task: task,
          scannedToken: raw,
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (_) => context.pop(),
      onFailure: (e) => setState(() { _handling = false; _error = e.message; }),
    );
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(handoffProvider(widget.handoffId));
    return Scaffold(
      appBar: const NxTopBar(title: 'Scan handoff QR'),
      body: AsyncValueWidget<HandoffTask>(
        value: async,
        data: (task) => Column(
          children: [
            Expanded(child: MobileScanner(onDetect: (c) => _onDetect(c, task))),
            Padding(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Column(
                children: [
                  Text('Scan courier handoff QR for order ${task.orderId}',
                      style: NxTypography.bodyMd, textAlign: TextAlign.center),
                  if (_error != null) ...[
                    const SizedBox(height: NxSpacing.s2),
                    Text(_error!, style: NxTypography.captionMd.copyWith(color: context.nxColors.danger)),
                  ],
                  if (_handling) ...[
                    const SizedBox(height: NxSpacing.s2),
                    const NxSpinner(),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
