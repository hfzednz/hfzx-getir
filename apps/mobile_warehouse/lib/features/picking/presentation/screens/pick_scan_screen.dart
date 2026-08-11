import 'package:collection/collection.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/pick_task.dart';
import '../providers/picking_providers.dart';

class PickScanScreen extends ConsumerStatefulWidget {
  const PickScanScreen({super.key, required this.taskId});

  final String taskId;

  @override
  ConsumerState<PickScanScreen> createState() => _PickScanScreenState();
}

class _PickScanScreenState extends ConsumerState<PickScanScreen> {
  bool _handling = false;
  String? _error;

  Future<void> _onDetect(BarcodeCapture capture, PickTask task) async {
    if (_handling) return;
    final raw = capture.barcodes.firstOrNull?.rawValue;
    if (raw == null || raw.isEmpty) return;

    final line = task.pathOrderedLines.firstWhereOrNull((l) => !l.isComplete);
    if (line == null) {
      setState(() => _error = 'All lines complete');
      return;
    }

    setState(() {
      _handling = true;
      _error = null;
    });

    final result = await ref.read(pickingActionsProvider).scanLine(
          task: task,
          line: line,
          barcode: raw,
        );

    if (!mounted) return;
    result.fold(
      onSuccess: (_) => context.pop(),
      onFailure: (e) {
        setState(() {
          _handling = false;
          _error = e.message;
        });
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(pickTaskProvider(widget.taskId));

    return Scaffold(
      appBar: const NxTopBar(title: 'Scan barcode'),
      body: AsyncValueWidget<PickTask>(
        value: async,
        data: (task) {
          final next =
              task.pathOrderedLines.firstWhereOrNull((l) => !l.isComplete);
          return Column(
            children: [
              Expanded(
                child: MobileScanner(
                  onDetect: (capture) => _onDetect(capture, task),
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(NxSpacing.s4),
                child: Column(
                  children: [
                    Text(
                      next == null
                          ? 'All lines picked'
                          : 'Scan ${next.sku} @ bin ${next.bin}',
                      style: NxTypography.bodyMd,
                      textAlign: TextAlign.center,
                    ),
                    if (_error != null) ...[
                      const SizedBox(height: NxSpacing.s2),
                      Text(
                        _error!,
                        style: NxTypography.captionMd
                            .copyWith(color: context.nxColors.danger),
                      ),
                    ],
                    if (_handling) ...[
                      const SizedBox(height: NxSpacing.s2),
                      const NxSpinner(),
                    ],
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}
