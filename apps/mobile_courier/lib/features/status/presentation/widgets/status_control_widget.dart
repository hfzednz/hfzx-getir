import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../domain/entities/duty_status.dart';
import '../providers/duty_controller.dart';

/// Compact duty control for home / shell glanceable status.
class StatusControlWidget extends ConsumerWidget {
  const StatusControlWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final duty = ref.watch(dutyControllerProvider);
    final colors = context.nxColors;
    final controller = ref.read(dutyControllerProvider.notifier);

    return NxCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Row(
            children: [
              _StatusDot(status: duty.status),
              const SizedBox(width: NxSpacing.s2),
              Expanded(
                child: Text(
                  _label(duty.status),
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                ),
              ),
              if (duty.isLoading) const NxSpinner(size: NxSpinnerSize.sm),
            ],
          ),
          const SizedBox(height: NxSpacing.s3),
          Wrap(
            spacing: NxSpacing.s2,
            runSpacing: NxSpacing.s2,
            children: [
              _ChipAction(
                label: duty.status == DutyStatus.offline ? 'Go online' : 'Go offline',
                onPressed: duty.isLoading
                    ? null
                    : () => controller.toggleOnline(),
              ),
              _ChipAction(
                label: 'Busy',
                selected: duty.status == DutyStatus.busy,
                onPressed: duty.isLoading
                    ? null
                    : () => controller.setStatus(DutyStatus.busy),
              ),
              _ChipAction(
                label: 'On break',
                selected: duty.status == DutyStatus.onBreak,
                onPressed: duty.isLoading
                    ? null
                    : () => controller.setStatus(DutyStatus.onBreak),
              ),
              _ChipAction(
                label: 'Emergency',
                destructive: true,
                selected: duty.status == DutyStatus.emergency,
                onPressed: duty.isLoading
                    ? null
                    : () => controller.setStatus(DutyStatus.emergency),
              ),
            ],
          ),
          if (duty.error != null) ...[
            const SizedBox(height: NxSpacing.s2),
            Text(
              duty.error!,
              style: NxTypography.captionMd.copyWith(color: colors.danger),
            ),
          ],
        ],
      ),
    );
  }

  String _label(DutyStatus status) => switch (status) {
        DutyStatus.offline => 'Offline',
        DutyStatus.online => 'Online',
        DutyStatus.busy => 'Busy',
        DutyStatus.onBreak => 'On break',
        DutyStatus.emergency => 'Emergency',
      };
}

class _StatusDot extends StatelessWidget {
  const _StatusDot({required this.status});
  final DutyStatus status;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final color = switch (status) {
      DutyStatus.online => colors.success,
      DutyStatus.busy => colors.warning,
      DutyStatus.onBreak => colors.info,
      DutyStatus.emergency => colors.danger,
      DutyStatus.offline => colors.textTertiary,
    };
    return Container(
      width: 10,
      height: 10,
      decoration: BoxDecoration(color: color, shape: BoxShape.circle),
    );
  }
}

class _ChipAction extends StatelessWidget {
  const _ChipAction({
    required this.label,
    this.onPressed,
    this.selected = false,
    this.destructive = false,
  });

  final String label;
  final VoidCallback? onPressed;
  final bool selected;
  final bool destructive;

  @override
  Widget build(BuildContext context) {
    return NxButton(
      label: label,
      size: NxButtonSize.sm,
      variant: destructive
          ? NxButtonVariant.destructive
          : selected
              ? NxButtonVariant.primary
              : NxButtonVariant.secondary,
      onPressed: onPressed,
    );
  }
}
