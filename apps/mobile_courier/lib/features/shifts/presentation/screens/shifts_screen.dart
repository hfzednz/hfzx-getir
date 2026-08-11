import 'package:collection/collection.dart';
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
    final async = ref.watch(shiftsProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Shifts'),
      body: AsyncValueWidget<List<CourierShift>>(
        value: async,
        data: (shifts) {
          final active = shifts.where((s) => s.isActive).firstOrNull;
          return ListView(
            padding: const EdgeInsets.all(NxSpacing.s4),
            children: [
              if (active != null) ...[
                NxCard(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Active shift',
                        style: NxTypography.titleSm
                            .copyWith(color: colors.textPrimary),
                      ),
                      Text(
                        'Worked ${active.workedMinutes} / ${active.plannedMinutes} min',
                        style: NxTypography.bodyMd
                            .copyWith(color: colors.textSecondary),
                      ),
                      if (active.overtime)
                        Text(
                          'Overtime',
                          style: NxTypography.captionMd
                              .copyWith(color: colors.warning),
                        ),
                      const SizedBox(height: NxSpacing.s3),
                      Row(
                        children: [
                          Expanded(
                            child: NxButton(
                              label: active.onBreak ? 'End break' : 'Break',
                              variant: NxButtonVariant.secondary,
                              onPressed: () {
                                final actions = ref.read(shiftActionsProvider);
                                if (active.onBreak) {
                                  actions.endBreak(active.id);
                                } else {
                                  actions.startBreak(active.id);
                                }
                              },
                            ),
                          ),
                          const SizedBox(width: NxSpacing.s2),
                          Expanded(
                            child: NxButton(
                              label: 'End shift',
                              variant: NxButtonVariant.destructive,
                              onPressed: () =>
                                  ref.read(shiftActionsProvider).end(active.id),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ] else
                NxButton(
                  label: 'Start shift',
                  expand: true,
                  onPressed: () => ref.read(shiftActionsProvider).start(),
                ),
              const SizedBox(height: NxSpacing.s4),
              Text(
                'History',
                style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
              ),
              const SizedBox(height: NxSpacing.s2),
              if (shifts.isEmpty)
                const NxEmptyState(
                  title: 'Nothing here yet',
                  body: 'Check back soon.',
                )
              else
                ...shifts.map(
                  (s) => Padding(
                    padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                    child: NxCard(
                      child: Text(
                        '${s.startedAt ?? '—'} → ${s.endedAt ?? 'active'}'
                        '${s.overtime ? ' · OT' : ''}',
                        style: NxTypography.bodySm,
                      ),
                    ),
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}
