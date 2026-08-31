import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class HelpScreen extends StatelessWidget {
  const HelpScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final faqs = <(String, String)>[
      (l10n.helpFaqPlaceQ, l10n.helpFaqPlaceA),
      (l10n.helpFaqEtaQ, l10n.helpFaqEtaA),
      (l10n.helpFaqTrackQ, l10n.helpFaqTrackA),
      (l10n.helpFaqRefundQ, l10n.helpFaqRefundA),
      (l10n.helpFaqScheduleQ, l10n.helpFaqScheduleA),
      (l10n.helpFaqCouponQ, l10n.helpFaqCouponA),
    ];

    return Scaffold(
      appBar: NxTopBar(title: l10n.helpTitle),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text(l10n.frequentlyAskedQuestions, style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s3),
          ...faqs.map(
            (faq) => Padding(
              padding: const EdgeInsets.only(bottom: NxSpacing.s2),
              child: NxCard(
                child: ExpansionTile(
                  title: Text(faq.$1),
                  childrenPadding: const EdgeInsets.fromLTRB(
                    NxSpacing.s4,
                    0,
                    NxSpacing.s4,
                    NxSpacing.s4,
                  ),
                  children: [
                    Align(
                      alignment: Alignment.centerLeft,
                      child: Text(faq.$2, style: NxTypography.bodyMd),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const SizedBox(height: NxSpacing.s4),
          NxButton(
            label: l10n.supportTitle,
            expand: true,
            onPressed: () => context.push(RouteNames.support),
          ),
          const SizedBox(height: NxSpacing.s3),
          NxButton(
            label: l10n.supportAssistant,
            expand: true,
            variant: NxButtonVariant.secondary,
            onPressed: () => context.push(RouteNames.supportAssistant),
          ),
        ],
      ),
    );
  }
}
