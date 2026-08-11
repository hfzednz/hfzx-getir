import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../domain/entities/support_entity.dart';
import '../providers/support_providers.dart';

class SupportAssistantScreen extends ConsumerStatefulWidget {
  const SupportAssistantScreen({super.key, this.orderId});

  final String? orderId;

  @override
  ConsumerState<SupportAssistantScreen> createState() => _SupportAssistantScreenState();
}

class _SupportAssistantScreenState extends ConsumerState<SupportAssistantScreen> {
  final _controller = TextEditingController();

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final messages = ref.watch(supportAssistantMessagesProvider);
    final sendState = ref.watch(supportAssistantSendProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.chatSupport),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: messages.length,
              itemBuilder: (context, index) => _AssistantBubble(message: messages[index]),
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _controller,
                      decoration: const InputDecoration(hintText: 'Ask anything…', border: OutlineInputBorder()),
                      onSubmitted: (_) => _send(),
                    ),
                  ),
                  const SizedBox(width: NxSpacing.s2),
                  IconButton(
                    icon: sendState.isLoading
                        ? const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Icon(Icons.send),
                    onPressed: sendState.isLoading ? null : _send,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  void _send() {
    final text = _controller.text.trim();
    if (text.isEmpty) return;
    _controller.clear();
    ref.read(supportAssistantSendProvider.notifier).send(text, orderId: widget.orderId);
  }
}

class _AssistantBubble extends StatelessWidget {
  const _AssistantBubble({required this.message});

  final SupportAssistantMessage message;

  @override
  Widget build(BuildContext context) {
    final isUser = !message.isAssistant;
    return Align(
      alignment: isUser ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        margin: const EdgeInsets.only(bottom: NxSpacing.s2),
        padding: const EdgeInsets.all(NxSpacing.s3),
        constraints: BoxConstraints(maxWidth: MediaQuery.sizeOf(context).width * 0.8),
        decoration: BoxDecoration(
          color: isUser ? context.nxColors.bgBrand : context.nxColors.bgSurfaceRaised,
          borderRadius: BorderRadius.circular(NxRadius.md),
        ),
        child: Text(message.content),
      ),
    );
  }
}
