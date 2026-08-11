import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../providers/inventory_providers.dart';

class CycleCountScreen extends ConsumerStatefulWidget {
  const CycleCountScreen({super.key});

  @override
  ConsumerState<CycleCountScreen> createState() => _CycleCountScreenState();
}

class _CycleCountScreenState extends ConsumerState<CycleCountScreen> {
  String? _sessionId;
  String? _status;
  bool _loading = false;

  Future<void> _start() async {
    setState(() => _loading = true);
    final r = await ref.read(inventoryActionsProvider).startCycleCount();
    if (!mounted) return;
    r.fold(
      onSuccess: (s) => setState(() {
        _sessionId = s.id;
        _status = s.status;
        _loading = false;
      }),
      onFailure: (e) {
        setState(() => _loading = false);
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.message)));
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Cycle count'),
      body: Padding(
        padding: const EdgeInsets.all(NxSpacing.s3),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            if (_sessionId == null)
              NxButton(
                label: 'Start session',
                expand: true,
                loading: _loading,
                onPressed: _start,
              )
            else ...[
              NxCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Session $_sessionId', style: NxTypography.titleSm),
                    Text('Status: ${_status ?? '-'}', style: NxTypography.captionMd),
                  ],
                ),
              ),
              const SizedBox(height: NxSpacing.s3),
              Text(
                'Scan bins and submit counts from the scanner workflow. '
                'Use adjust screen for variance posting.',
                style: NxTypography.bodySm,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
