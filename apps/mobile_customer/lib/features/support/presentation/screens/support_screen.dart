import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../routing/route_names.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/support_providers.dart';

class SupportScreen extends ConsumerWidget {
  const SupportScreen({super.key, this.orderId});

  final String? orderId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final faqAsync = ref.watch(supportFaqProvider);
    final ticketsAsync = ref.watch(supportTicketsProvider);

    return Scaffold(
      appBar: NxTopBar(
        title: l10n.supportTitle,
        actions: [
          IconButton(
            icon: const Icon(Icons.smart_toy_outlined),
            tooltip: l10n.chatSupport,
            onPressed: () => context.push('${RouteNames.supportAssistant}${orderId != null ? '?orderId=$orderId' : ''}'),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateTicket(context, ref),
        label: Text(l10n.newTicket),
        icon: const Icon(Icons.add),
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(supportFaqProvider);
          ref.invalidate(supportTicketsProvider);
        },
        child: CustomScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          slivers: [
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(NxSpacing.s4),
                child: NxButton(
                  label: l10n.chatSupport,
                  variant: NxButtonVariant.secondary,
                  onPressed: () => context.push(RouteNames.supportAssistant),
                ),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(NxSpacing.s4, 0, NxSpacing.s4, NxSpacing.s2),
                child: Text(l10n.faqTitle, style: NxTypography.headlineSm),
              ),
            ),
            AsyncValueWidget(
              value: faqAsync,
              data: (faqs) => SliverList.separated(
                itemCount: faqs.length,
                separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
                itemBuilder: (context, index) => Padding(
                  padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                  child: NxCard(
                    child: ExpansionTile(
                      title: Text(faqs[index].question),
                      children: [Padding(padding: const EdgeInsets.all(NxSpacing.s4), child: Text(faqs[index].answer))],
                    ),
                  ),
                ),
              ),
              error: (e, _) => SliverToBoxAdapter(
                child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(supportFaqProvider)),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(NxSpacing.s4, NxSpacing.s4, NxSpacing.s4, NxSpacing.s2),
                child: Text(l10n.yourTickets, style: NxTypography.headlineSm),
              ),
            ),
            ticketsAsync.when(
              data: (tickets) => SliverList.separated(
                itemCount: tickets.length,
                separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
                itemBuilder: (context, index) {
                  final t = tickets[index];
                  return Padding(
                    padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                    child: NxCard(
                      child: ListTile(
                        title: Text(t.subject),
                        subtitle: Text(t.status.name),
                        trailing: t.liveChatUrl != null ? const Icon(Icons.chat) : const Icon(Icons.chevron_right),
                        onTap: () {
                          if (t.liveChatUrl != null) {
                            launchUrl(Uri.parse(t.liveChatUrl!), mode: LaunchMode.externalApplication);
                          } else {
                            context.push('${RouteNames.supportTicket}/${t.id}');
                          }
                        },
                      ),
                    ),
                  );
                },
              ),
              loading: () => const SliverToBoxAdapter(child: Center(child: CircularProgressIndicator())),
              error: (e, _) => SliverToBoxAdapter(
                child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(supportTicketsProvider)),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 80)),
          ],
        ),
      ),
    );
  }

  Future<void> _showCreateTicket(BuildContext context, WidgetRef ref) async {
    final l10n = AppLocalizations.of(context);
    final subjectController = TextEditingController();
    final bodyController = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.createSupportTicket),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(controller: subjectController, decoration: InputDecoration(labelText: l10n.subject)),
            TextField(controller: bodyController, decoration: InputDecoration(labelText: l10n.message), maxLines: 3),
          ],
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: Text(l10n.cancel)),
          FilledButton(onPressed: () => Navigator.pop(ctx, true), child: Text(l10n.submit)),
        ],
      ),
    );
    if (ok == true && context.mounted) {
      await ref.read(supportTicketCreateProvider.notifier).create(
            subject: subjectController.text,
            body: bodyController.text,
            orderId: orderId,
          );
    }
  }
}
