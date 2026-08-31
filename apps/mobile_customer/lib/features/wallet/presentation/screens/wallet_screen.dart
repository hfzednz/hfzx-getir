import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/errors/error_copy.dart';
import '../../../../shared/utils/formatters.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/wallet_entity.dart';
import '../providers/wallet_providers.dart';

class WalletScreen extends ConsumerWidget {
  const WalletScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final accountAsync = ref.watch(walletAccountProvider);
    final transactionsAsync = ref.watch(walletTransactionsProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.walletTitle),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(walletAccountProvider);
          ref.invalidate(walletTransactionsProvider);
        },
        child: AsyncValueWidget(
          value: accountAsync,
          data: (account) => CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            slivers: [
              SliverToBoxAdapter(child: _BalanceCard(account: account, onTopUp: () => _showTopUp(context, ref, account))),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, NxSpacing.s4, NxSpacing.s4, NxSpacing.s2),
                  child: Text(l10n.transactionHistory, style: NxTypography.headlineSm),
                ),
              ),
              AsyncValueWidget(
                value: transactionsAsync,
                data: (txns) {
                  if (txns.isEmpty) {
                    return SliverFillRemaining(
                      hasScrollBody: false,
                      child: NxEmptyState(title: l10n.emptyTitle, body: l10n.emptyMessage),
                    );
                  }
                  return SliverList.separated(
                    itemCount: txns.length,
                    separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
                    itemBuilder: (context, index) => Padding(
                      padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                      child: _TransactionTile(transaction: txns[index]),
                    ),
                  );
                },
                error: (e, _) => SliverToBoxAdapter(
                  child: ErrorView(message: localizedCustomerError(context, e), onRetry: () => ref.invalidate(walletTransactionsProvider)),
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: NxSpacing.s6)),
            ],
          ),
          error: (e, _) => ErrorView(message: localizedCustomerError(context, e), onRetry: () => ref.invalidate(walletAccountProvider)),
        ),
      ),
    );
  }

  Future<void> _showTopUp(BuildContext context, WidgetRef ref, WalletAccount account) async {
    if (!account.topUpEnabled) return;
    final l10n = AppLocalizations.of(context);
    final controller = TextEditingController();
    final amount = await showDialog<int>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(l10n.topUpWallet),
        content: TextField(
          controller: controller,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          decoration: InputDecoration(labelText: l10n.amountTry),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(l10n.cancel)),
          FilledButton(
            onPressed: () {
              final major = double.tryParse(controller.text.replaceAll(',', '.'));
              if (major == null || major <= 0) return;
              Navigator.pop(ctx, (major * 100).round());
            },
            child: Text(l10n.topUp),
          ),
        ],
      ),
    );
    if (amount != null && context.mounted) {
      await ref.read(walletTopUpControllerProvider.notifier).topUp(amountMinor: amount, currency: account.currency);
    }
  }
}

class _BalanceCard extends StatelessWidget {
  const _BalanceCard({required this.account, required this.onTopUp});

  final WalletAccount account;
  final VoidCallback onTopUp;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(NxSpacing.s4),
      child: NxCard(
        child: Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(l10n.availableBalance, style: NxTypography.captionMd),
              const SizedBox(height: NxSpacing.s1),
              Text(account.balance.format(), style: NxTypography.displayLg),
              if (account.pendingMinor > 0) ...[
                const SizedBox(height: NxSpacing.s2),
                Text(
                  '${l10n.pending}: ${Money(minorUnits: account.pendingMinor, currency: account.currency).format()}',
                  style: NxTypography.captionMd,
                ),
              ],
              const SizedBox(height: NxSpacing.s3),
              Wrap(
                spacing: NxSpacing.s2,
                runSpacing: NxSpacing.s2,
                children: [
                  if (account.cashbackMinor > 0)
                    NxChip(
                      label:
                          '${l10n.cashback} ${Money(minorUnits: account.cashbackMinor, currency: account.currency).format()}',
                    ),
                  if (account.promoCreditMinor > 0)
                    NxChip(
                      label:
                          '${l10n.promoCredit} ${Money(minorUnits: account.promoCreditMinor, currency: account.currency).format()}',
                    ),
                ],
              ),
              if (account.topUpEnabled) ...[
                const SizedBox(height: NxSpacing.s4),
                NxButton(label: l10n.topUp, onPressed: onTopUp, variant: NxButtonVariant.secondary),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _TransactionTile extends StatelessWidget {
  const _TransactionTile({required this.transaction});

  final WalletTransaction transaction;

  @override
  Widget build(BuildContext context) {
    final sign = transaction.isCredit ? '+' : '−';
    return NxCard(
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor:
              transaction.isCredit ? context.nxColors.successSurface : context.nxColors.bgSurfaceRaised,
          child: Icon(
            transaction.isCredit ? Icons.arrow_downward : Icons.arrow_upward,
            size: 18,
            color: transaction.isCredit ? context.nxColors.success : context.nxColors.textTertiary,
          ),
        ),
        title: Text(transaction.description.isNotEmpty ? transaction.description : transaction.type.label),
        subtitle: Text([
          transaction.type.label,
          if (transaction.createdAt != null) Formatters.dateTime(transaction.createdAt!),
        ].join(' · ')),
        trailing: Text(
          '$sign${transaction.amount.format()}',
          style: NxTypography.bodyMd.copyWith(
            color: transaction.isCredit ? context.nxColors.success : context.nxColors.textPrimary,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}
