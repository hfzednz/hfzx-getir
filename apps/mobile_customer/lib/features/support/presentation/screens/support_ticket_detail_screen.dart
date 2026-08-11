import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../shared/utils/formatters.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/support_entity.dart';
import '../providers/support_providers.dart';

class SupportTicketDetailScreen extends ConsumerWidget {
  const SupportTicketDetailScreen({super.key, required this.ticketId});

  final String ticketId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final ticketAsync = ref.watch(supportTicketProvider(ticketId));

    return Scaffold(
      appBar: NxTopBar(title: 'Ticket'),
      body: AsyncValueWidget(
        value: ticketAsync,
        data: (ticket) => Column(
          children: [
            if (ticket.liveChatUrl != null)
              Padding(
                padding: const EdgeInsets.all(NxSpacing.s4),
                child: NxButton(
                  label: 'Open live chat',
                  onPressed: () => launchUrl(Uri.parse(ticket.liveChatUrl!), mode: LaunchMode.externalApplication),
                ),
              ),
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.all(NxSpacing.s4),
                itemCount: ticket.messages.length,
                itemBuilder: (context, index) => _MessageBubble(message: ticket.messages[index]),
              ),
            ),
          ],
        ),
        error: (e, _) => ErrorView(message: e.toString(), onRetry: () => ref.invalidate(supportTicketProvider(ticketId))),
      ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  const _MessageBubble({required this.message});

  final SupportTicketMessage message;

  @override
  Widget build(BuildContext context) {
    final align = message.isFromUser ? Alignment.centerRight : Alignment.centerLeft;
    final color = message.isFromUser ? context.nxColors.bgBrand : context.nxColors.bgSurfaceRaised;
    return Align(
      alignment: align,
      child: Container(
        margin: const EdgeInsets.only(bottom: NxSpacing.s2),
        padding: const EdgeInsets.all(NxSpacing.s3),
        decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(NxRadius.md)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(message.body),
            if (message.createdAt != null)
              Text(Formatters.dateTime(message.createdAt!), style: NxTypography.captionMd),
          ],
        ),
      ),
    );
  }
}
