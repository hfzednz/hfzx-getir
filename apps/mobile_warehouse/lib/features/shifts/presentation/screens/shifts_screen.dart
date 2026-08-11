import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/shift_entity.dart';
import '../providers/shifts_providers.dart';

class ShiftsScreen extends ConsumerWidget {
  const ShiftsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final current = ref.watch(currentShiftProvider);
    final attendance = ref.watch(attendanceProvider);
    final actions = ref.read(shiftsActionsProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Shifts'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s3),
        children: [
          AsyncValueWidget<WarehouseShift?>(
            value: current,
            data: (shift) => NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(shift == null ? 'Not clocked in' : 'Shift ${shift.id}', style: NxTypography.titleSm),
                  if (shift != null)
                    Text('Status: ${shift.status}${shift.onBreak ? ' (break)' : ''}', style: NxTypography.captionMd),
                  const SizedBox(height: NxSpacing.s3),
                  Wrap(
                    spacing: NxSpacing.s2,
                    runSpacing: NxSpacing.s2,
                    children: [
                      NxButton(label: 'Clock in', size: NxButtonSize.sm, onPressed: actions.clockIn),
                      NxButton(label: 'Clock out', size: NxButtonSize.sm, variant: NxButtonVariant.secondary, onPressed: actions.clockOut),
                      NxButton(label: 'Start break', size: NxButtonSize.sm, variant: NxButtonVariant.tertiary, onPressed: actions.startBreak),
                      NxButton(label: 'End break', size: NxButtonSize.sm, variant: NxButtonVariant.tertiary, onPressed: actions.endBreak),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: NxSpacing.s4),
          Text('Attendance', style: NxTypography.titleSm),
          const SizedBox(height: NxSpacing.s2),
          AsyncValueWidget<List<WarehouseShift>>(
            value: attendance,
            data: (items) => Column(
              children: items.map((s) => Padding(
                padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                child: NxCard(child: Text('${s.id} · ${s.status}', style: NxTypography.bodySm)),
              )).toList(),
            ),
          ),
        ],
      ),
    );
  }
}
