import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/business_rules/delivery_rules.dart';
import '../providers/deliveries_providers.dart';

class FailedDeliveryScreen extends ConsumerStatefulWidget {
  const FailedDeliveryScreen({super.key, required this.deliveryId});

  final String deliveryId;

  @override
  ConsumerState<FailedDeliveryScreen> createState() =>
      _FailedDeliveryScreenState();
}

class _FailedDeliveryScreenState extends ConsumerState<FailedDeliveryScreen> {
  String? _reason;
  final _noteController = TextEditingController();
  bool _submitting = false;
  String? _error;

  static const _labels = {
    'customer_unavailable': 'Customer unavailable',
    'wrong_address': 'Wrong address',
    'refused': 'Customer refused',
    'unsafe_location': 'Unsafe location',
    'damaged_goods': 'Damaged goods',
    'other': 'Other',
  };

  @override
  void dispose() {
    _noteController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() {
      _submitting = true;
      _error = null;
    });
    final updated = await ref.read(deliveryActionsProvider).markFailed(
          id: widget.deliveryId,
          reasonCode: _reason ?? '',
          note: _noteController.text.trim().isEmpty
              ? null
              : _noteController.text.trim(),
        );
    if (!mounted) return;
    if (updated == null) {
      setState(() {
        _submitting = false;
        _error = 'Invalid failure reason';
      });
      return;
    }
    context.pop();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Mark as failed'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          ...DeliveryRules.failureReasonCodes.map((code) {
            return RadioListTile<String>(
              value: code,
              groupValue: _reason,
              title: Text(_labels[code] ?? code),
              onChanged: (v) => setState(() => _reason = v),
            );
          }),
          const SizedBox(height: NxSpacing.s3),
          TextField(
            controller: _noteController,
            maxLines: 3,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              hintText: 'Optional note',
            ),
          ),
          if (_error != null) ...[
            const SizedBox(height: NxSpacing.s2),
            Text(
              _error!,
              style: NxTypography.captionMd
                  .copyWith(color: context.nxColors.danger),
            ),
          ],
          const SizedBox(height: NxSpacing.s4),
          NxButton(
            label: 'Submit',
            variant: NxButtonVariant.destructive,
            expand: true,
            loading: _submitting,
            disabled: _reason == null,
            onPressed: _submitting || _reason == null ? null : _submit,
          ),
        ],
      ),
    );
  }
}
