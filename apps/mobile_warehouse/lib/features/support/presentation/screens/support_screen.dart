import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/support_entity.dart';
import '../providers/support_providers.dart';

class SupportScreen extends ConsumerStatefulWidget {
  const SupportScreen({super.key});
  @override
  ConsumerState<SupportScreen> createState() => _SupportScreenState();
}

class _SupportScreenState extends ConsumerState<SupportScreen> {
  final _subject = TextEditingController();
  final _body = TextEditingController();
  @override
  void dispose() {
    _subject.dispose();
    _body.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final async = ref.watch(supportTicketsProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Support'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s3),
        children: [
          TextField(controller: _subject, decoration: const InputDecoration(labelText: 'Subject')),
          TextField(controller: _body, decoration: const InputDecoration(labelText: 'Details'), maxLines: 3),
          const SizedBox(height: NxSpacing.s2),
          NxButton(
            label: 'Create ticket',
            expand: true,
            onPressed: () => ref.read(supportActionsProvider).create(
                  subject: _subject.text.trim(),
                  body: _body.text.trim(),
                ),
          ),
          const SizedBox(height: NxSpacing.s4),
          Text('Tickets', style: NxTypography.titleSm),
          const SizedBox(height: NxSpacing.s2),
          AsyncValueWidget<List<SupportTicket>>(
            value: async,
            data: (items) => Column(
              children: items.map((t) => Padding(
                padding: const EdgeInsets.only(bottom: NxSpacing.s2),
                child: NxCard(child: Text('${t.subject} · ${t.status}', style: NxTypography.bodySm)),
              )).toList(),
            ),
          ),
        ],
      ),
    );
  }
}
