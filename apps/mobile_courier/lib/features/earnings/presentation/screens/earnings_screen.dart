import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/earnings_entity.dart';
import '../providers/earnings_providers.dart';

class EarningsScreen extends ConsumerWidget {
  const EarningsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final period = ref.watch(earningsPeriodProvider);
    final async = ref.watch(earningsProvider);
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Earnings'),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Row(
              children: [
                for (final p in EarningsPeriod.values) ...[
                  Expanded(
                    child: NxButton(
                      label: p.name,
                      size: NxButtonSize.sm,
                      variant: period == p
                          ? NxButtonVariant.primary
                          : NxButtonVariant.secondary,
                      onPressed: () =>
                          ref.read(earningsPeriodProvider.notifier).state = p,
                    ),
                  ),
                  if (p != EarningsPeriod.monthly)
                    const SizedBox(width: NxSpacing.s2),
                ],
              ],
            ),
          ),
          Expanded(
            child: AsyncValueWidget<EarningsSnapshot>(
              value: async,
              data: (snap) {
                final b = snap.breakdown;
                final money = (int minor) =>
                    Money(minorUnits: minor, currency: b.currency).format();
                return ListView(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                  children: [
                    NxCard(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            "Today's earnings".replaceFirst(
                              "Today's",
                              period == EarningsPeriod.daily
                                  ? "Today's"
                                  : period == EarningsPeriod.weekly
                                      ? "Week's"
                                      : "Month's",
                            ),
                            style: NxTypography.captionMd
                                .copyWith(color: colors.textSecondary),
                          ),
                          Text(
                            money(b.netMinor),
                            style: NxTypography.dashKpi
                                .copyWith(color: colors.textPrimary),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: NxSpacing.s3),
                    _row('Base', money(b.baseMinor), colors),
                    _row('Tips', money(b.tipsMinor), colors),
                    _row('Bonuses', money(b.bonusesMinor), colors),
                    _row('Penalties', '-${money(b.penaltiesMinor)}', colors),
                    const SizedBox(height: NxSpacing.s4),
                    Text(
                      'Payout history',
                      style: NxTypography.titleSm
                          .copyWith(color: colors.textPrimary),
                    ),
                    const SizedBox(height: NxSpacing.s2),
                    if (snap.payouts.isEmpty)
                      Text(
                        'No payouts yet',
                        style: NxTypography.bodySm
                            .copyWith(color: colors.textSecondary),
                      )
                    else
                      ...snap.payouts.map(
                        (p) => Padding(
                          padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                          child: NxCard(
                            child: Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    p.status,
                                    style: NxTypography.bodySm,
                                  ),
                                ),
                                Text(
                                  Money(
                                    minorUnits: p.amountMinor,
                                    currency: p.currency,
                                  ).format(),
                                  style: NxTypography.priceSm,
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _row(String label, String value, NxColorRoles colors) {
    return Padding(
      padding: const EdgeInsets.only(bottom: NxSpacing.s2),
      child: NxCard(
        child: Row(
          children: [
            Expanded(
              child: Text(
                label,
                style: NxTypography.bodyMd.copyWith(color: colors.textSecondary),
              ),
            ),
            Text(
              value,
              style: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
            ),
          ],
        ),
      ),
    );
  }
}
