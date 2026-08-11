import 'package:collection/collection.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/delivery_job.dart';
import '../providers/deliveries_providers.dart';

class PickupScanScreen extends ConsumerStatefulWidget {
  const PickupScanScreen({super.key, required this.deliveryId});

  final String deliveryId;

  @override
  ConsumerState<PickupScanScreen> createState() => _PickupScanScreenState();
}

class _PickupScanScreenState extends ConsumerState<PickupScanScreen> {
  bool _handling = false;
  String? _error;

  Future<void> _onDetect(BarcodeCapture capture, DeliveryJob job) async {
    if (_handling) return;
    final raw = capture.barcodes.firstOrNull?.rawValue;
    if (raw == null || raw.isEmpty) return;

    setState(() {
      _handling = true;
      _error = null;
    });

    final updated = await ref.read(deliveryActionsProvider).confirmPickup(
          job: job,
          scannedToken: raw,
        );

    if (!mounted) return;
    if (updated == null) {
      setState(() {
        _handling = false;
        _error = 'QR token does not match warehouse handoff';
      });
      return;
    }
    context.pop();
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(deliveryDetailProvider(widget.deliveryId));

    return Scaffold(
      appBar: const NxTopBar(title: 'Scan QR'),
      body: AsyncValueWidget<DeliveryJob>(
        value: async,
        data: (job) => Column(
          children: [
            Expanded(
              child: MobileScanner(
                onDetect: (capture) => _onDetect(capture, job),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Column(
                children: [
                  Text(
                    'Scan warehouse handoff QR for ${job.storeName}',
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
                  const SizedBox(height: NxSpacing.s3),
                  NxButton(
                    label: 'Confirm pickup',
                    expand: true,
                    disabled: _handling,
                    onPressed: () async {
                      // Manual confirm uses expected token when scanner unavailable.
                      final token = job.handoffToken;
                      if (token == null) {
                        setState(() => _error = 'Missing handoff token');
                        return;
                      }
                      setState(() => _handling = true);
                      final updated = await ref
                          .read(deliveryActionsProvider)
                          .confirmPickup(job: job, scannedToken: token);
                      if (!mounted) return;
                      if (updated == null) {
                        setState(() {
                          _handling = false;
                          _error = 'Pickup confirmation failed';
                        });
                        return;
                      }
                      context.pop();
                    },
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
