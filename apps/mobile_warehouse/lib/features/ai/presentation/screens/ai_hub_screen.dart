import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/ai_insights.dart';
import '../providers/ai_providers.dart';

class AiHubScreen extends ConsumerWidget {
  const AiHubScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(aiHubProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'AI assist'),
      body: AsyncValueWidget<AiHubInsights>(
        value: async,
        data: (hub) => ListView(
          padding: const EdgeInsets.all(NxSpacing.s3),
          children: [
            Text('Demand forecast', style: NxTypography.titleSm),
            const SizedBox(height: NxSpacing.s2),
            NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: hub.demandForecast.isEmpty
                    ? [Text('No forecast', style: NxTypography.bodySm)]
                    : hub.demandForecast
                        .map((e) => Text('• $e', style: NxTypography.bodySm))
                        .toList(),
              ),
            ),
            const SizedBox(height: NxSpacing.s3),
            Text('Pick path tip', style: NxTypography.titleSm),
            const SizedBox(height: NxSpacing.s2),
            NxCard(
              child: Text(
                hub.pickPathTip ?? 'No tip',
                style: NxTypography.bodySm,
              ),
            ),
            const SizedBox(height: NxSpacing.s3),
            Text('Restock suggestions', style: NxTypography.titleSm),
            const SizedBox(height: NxSpacing.s2),
            NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: hub.restockSuggestions.isEmpty
                    ? [Text('None', style: NxTypography.bodySm)]
                    : hub.restockSuggestions
                        .map((e) => Text('• $e', style: NxTypography.bodySm))
                        .toList(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
