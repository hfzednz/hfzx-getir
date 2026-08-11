import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/delivery_job.dart';
import '../providers/deliveries_providers.dart';

class PodScreen extends ConsumerStatefulWidget {
  const PodScreen({super.key, required this.deliveryId});

  final String deliveryId;

  @override
  ConsumerState<PodScreen> createState() => _PodScreenState();
}

class _PodScreenState extends ConsumerState<PodScreen> {
  final _otpController = TextEditingController();
  final _noteController = TextEditingController();
  String? _photoPath;
  bool _submitting = false;
  String? _error;

  @override
  void dispose() {
    _otpController.dispose();
    _noteController.dispose();
    super.dispose();
  }

  Future<void> _pickPhoto() async {
    final file = await ImagePicker().pickImage(
      source: ImageSource.camera,
      imageQuality: 80,
    );
    if (file != null) setState(() => _photoPath = file.path);
  }

  Future<void> _submit(DeliveryJob job) async {
    setState(() {
      _submitting = true;
      _error = null;
    });
    final updated = await ref.read(deliveryActionsProvider).submitPod(
          job: job,
          photoPath: _photoPath ?? '',
          otp: _otpController.text.trim().isEmpty
              ? null
              : _otpController.text.trim(),
          signatureNote: _noteController.text.trim().isEmpty
              ? null
              : _noteController.text.trim(),
        );
    if (!mounted) return;
    if (updated == null) {
      setState(() {
        _submitting = false;
        _error = 'Could not submit proof of delivery';
      });
      return;
    }
    context.pop();
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(deliveryDetailProvider(widget.deliveryId));

    return Scaffold(
      appBar: const NxTopBar(title: 'Proof of delivery'),
      body: AsyncValueWidget<DeliveryJob>(
        value: async,
        data: (job) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s4),
          children: [
            NxButton(
              label: _photoPath == null ? 'Take photo' : 'Retake photo',
              variant: NxButtonVariant.secondary,
              expand: true,
              onPressed: _pickPhoto,
            ),
            if (_photoPath != null) ...[
              const SizedBox(height: NxSpacing.s2),
              Text(
                'Photo captured',
                style: NxTypography.captionMd
                    .copyWith(color: context.nxColors.success),
              ),
            ],
            const SizedBox(height: NxSpacing.s4),
            if (job.otpRequired) ...[
              Text('OTP', style: NxTypography.captionMd),
              const SizedBox(height: NxSpacing.s1),
              NxOtpInput(
                length: 6,
                onCompleted: (v) => _otpController.text = v,
              ),
              const SizedBox(height: NxSpacing.s4),
            ],
            Text('Signature note', style: NxTypography.captionMd),
            const SizedBox(height: NxSpacing.s1),
            TextField(
              controller: _noteController,
              maxLines: 3,
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                hintText: 'Optional note / signature acknowledgement',
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
              label: 'Confirm delivery',
              expand: true,
              loading: _submitting,
              onPressed: _submitting ? null : () => _submit(job),
            ),
          ],
        ),
      ),
    );
  }
}
