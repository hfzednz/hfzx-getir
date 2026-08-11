import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/support_entity.dart';
import '../providers/support_providers.dart';

class SupportScreen extends ConsumerStatefulWidget {
  const SupportScreen({super.key});

  @override
  ConsumerState<SupportScreen> createState() => _SupportScreenState();
}

class _SupportScreenState extends ConsumerState<SupportScreen> {
  final _chatController = TextEditingController();
  final _incidentController = TextEditingController();
  int _tab = 0;

  @override
  void dispose() {
    _chatController.dispose();
    _incidentController.dispose();
    super.dispose();
  }

  Future<void> _sos() async {
    final tel = Uri.parse('tel:112');
    if (await canLaunchUrl(tel)) {
      await launchUrl(tel);
    }
    await ref.read(supportActionsProvider).reportIncident(
          type: 'sos',
          description: 'SOS triggered from courier app',
        );
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return Scaffold(
      appBar: const NxTopBar(title: 'Support'),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Row(
              children: [
                Expanded(
                  child: NxButton(
                    label: 'AI chat',
                    size: NxButtonSize.sm,
                    variant: _tab == 0
                        ? NxButtonVariant.primary
                        : NxButtonVariant.secondary,
                    onPressed: () => setState(() => _tab = 0),
                  ),
                ),
                const SizedBox(width: NxSpacing.s2),
                Expanded(
                  child: NxButton(
                    label: 'Tickets',
                    size: NxButtonSize.sm,
                    variant: _tab == 1
                        ? NxButtonVariant.primary
                        : NxButtonVariant.secondary,
                    onPressed: () => setState(() => _tab = 1),
                  ),
                ),
                const SizedBox(width: NxSpacing.s2),
                Expanded(
                  child: NxButton(
                    label: 'SOS',
                    size: NxButtonSize.sm,
                    variant: NxButtonVariant.destructive,
                    onPressed: _sos,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: _tab == 0 ? _buildChat(colors) : _buildTickets(),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                TextField(
                  controller: _incidentController,
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                    hintText: 'Incident report…',
                  ),
                ),
                const SizedBox(height: NxSpacing.s2),
                NxButton(
                  label: 'Report incident',
                  variant: NxButtonVariant.secondary,
                  onPressed: () async {
                    final text = _incidentController.text.trim();
                    if (text.isEmpty) return;
                    await ref.read(supportActionsProvider).reportIncident(
                          type: 'manual',
                          description: text,
                        );
                    _incidentController.clear();
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChat(NxColorRoles colors) {
    final async = ref.watch(supportChatProvider);
    return Column(
      children: [
        Expanded(
          child: AsyncValueWidget<List<ChatMessage>>(
            value: async,
            data: (messages) {
              if (messages.isEmpty) {
                return const NxEmptyState(
                  title: 'AI assistant',
                  body: 'Ask a question about deliveries or payouts.',
                );
              }
              return ListView.builder(
                padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                itemCount: messages.length,
                itemBuilder: (context, i) {
                  final m = messages[i];
                  return Align(
                    alignment: m.fromAssistant
                        ? Alignment.centerLeft
                        : Alignment.centerRight,
                    child: Container(
                      margin: const EdgeInsets.only(bottom: NxSpacing.s2),
                      padding: const EdgeInsets.all(NxSpacing.s3),
                      decoration: BoxDecoration(
                        color: m.fromAssistant
                            ? colors.bgSunken
                            : colors.bgBrand,
                        borderRadius: BorderRadius.circular(NxRadius.md),
                      ),
                      child: Text(
                        m.text,
                        style: NxTypography.bodySm.copyWith(
                          color: m.fromAssistant
                              ? colors.textPrimary
                              : colors.textOnBrand,
                        ),
                      ),
                    ),
                  );
                },
              );
            },
          ),
        ),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _chatController,
                  decoration: const InputDecoration(
                    border: OutlineInputBorder(),
                    hintText: 'Message…',
                  ),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.send),
                onPressed: () async {
                  final text = _chatController.text.trim();
                  if (text.isEmpty) return;
                  _chatController.clear();
                  await ref.read(supportActionsProvider).sendChat(text);
                },
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTickets() {
    final async = ref.watch(supportTicketsProvider);
    return AsyncValueWidget<List<SupportTicket>>(
      value: async,
      data: (tickets) {
        if (tickets.isEmpty) {
          return const NxEmptyState(
            title: 'Nothing here yet',
            body: 'Check back soon.',
          );
        }
        return ListView.separated(
          padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
          itemCount: tickets.length,
          separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
          itemBuilder: (context, i) {
            final t = tickets[i];
            return NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(t.subject, style: NxTypography.titleSm),
                  Text(t.status, style: NxTypography.captionMd),
                ],
              ),
            );
          },
        );
      },
    );
  }
}
