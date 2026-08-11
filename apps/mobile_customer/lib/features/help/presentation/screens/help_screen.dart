import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';

class HelpScreen extends StatelessWidget {
  const HelpScreen({super.key, this.id});

  final String? id;

  static const _faqs = <(String question, String answer)>[
    (
      'How do I place an order?',
      'Browse categories or search for products, add items to your cart, '
          'then checkout with a delivery address and payment method.',
    ),
    (
      'What are delivery times?',
      'ETA is shown on product and checkout screens based on your city and '
          'selected address. Most orders arrive within the estimated window.',
    ),
    (
      'How do I track my order?',
      'Open Orders from your account, select the order, then tap Track. '
          'You will see live status updates until delivery.',
    ),
    (
      'How do refunds work?',
      'If an item is missing or damaged, open the order and contact Support. '
          'Approved refunds return to your original payment method or wallet.',
    ),
    (
      'Can I schedule a delivery?',
      'Yes. During checkout choose Schedule delivery and pick an available slot '
          'for your address.',
    ),
    (
      'How do coupons and loyalty points work?',
      'Apply a coupon on checkout. Loyalty points accumulate on eligible orders '
          'and can be redeemed where shown in your wallet or loyalty screen.',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return Scaffold(
      appBar: NxTopBar(title: l10n.helpTitle),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text('Frequently asked questions', style: NxTypography.headlineSm),
          const SizedBox(height: NxSpacing.s3),
          ..._faqs.map(
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
            label: 'Support assistant',
            expand: true,
            variant: NxButtonVariant.secondary,
            onPressed: () => context.push(RouteNames.supportAssistant),
          ),
        ],
      ),
    );
  }
}
