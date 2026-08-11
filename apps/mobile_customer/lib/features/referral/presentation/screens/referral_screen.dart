import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/referral_entity.dart';
import '../providers/referral_providers.dart';

class ReferralScreen extends ConsumerStatefulWidget {
  const ReferralScreen({super.key});

  @override
  ConsumerState<ReferralScreen> createState() => _ReferralScreenState();
}

class _ReferralScreenState extends ConsumerState<ReferralScreen> {
  final _claimController = TextEditingController();

  @override
  void dispose() {
    _claimController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final infoAsync = ref.watch(referralInfoProvider);
    final invitesAsync = ref.watch(referralInvitesProvider);
    final claimState = ref.watch(referralClaimProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.referralTitle),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(referralInfoProvider);
          ref.invalidate(referralInvitesProvider);
        },
        child: AsyncValueWidget(
          value: infoAsync,
          data: (info) => CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            slivers: [
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.all(NxSpacing.s4),
                  child: NxCard(
                    child: Padding(
                      padding: const EdgeInsets.all(NxSpacing.s4),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(l10n.referralCode, style: NxTypography.captionMd),
                          const SizedBox(height: NxSpacing.s2),
                          Row(
                            children: [
                              Expanded(
                                child: Text(info.inviteCode, style: NxTypography.displayLg),
                              ),
                              IconButton(
                                icon: const Icon(Icons.copy),
                                onPressed: () => Clipboard.setData(ClipboardData(text: info.inviteCode)),
                              ),
                            ],
                          ),
                          const SizedBox(height: NxSpacing.s3),
                          Text(
                            '${info.successfulInvites}/${info.totalInvites} successful invites',
                            style: NxTypography.bodyMd,
                          ),
                          if (info.rewardMinor > 0) ...[
                            const SizedBox(height: NxSpacing.s1),
                            Text(
                              'Earned ${Money(minorUnits: info.rewardMinor, currency: info.currency).format()}',
                              style: NxTypography.captionMd,
                            ),
                          ],
                          const SizedBox(height: NxSpacing.s4),
                          NxButton(
                            label: l10n.shareReferral,
                            onPressed: () => shareReferralInvite(ref, info),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, 0, NxSpacing.s4, NxSpacing.s4),
                  child: NxCard(
                    child: Padding(
                      padding: const EdgeInsets.all(NxSpacing.s4),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Have an invite code?', style: NxTypography.headlineSm),
                          const SizedBox(height: NxSpacing.s3),
                          TextField(
                            controller: _claimController,
                            textCapitalization: TextCapitalization.characters,
                            decoration: const InputDecoration(
                              labelText: 'Invite code',
                              border: OutlineInputBorder(),
                            ),
                          ),
                          if (claimState.hasError) ...[
                            const SizedBox(height: NxSpacing.s2),
                            Text(
                              claimState.error.toString(),
                              style: NxTypography.captionMd.copyWith(
                                color: context.nxColors.danger,
                              ),
                            ),
                          ],
                          if (claimState.valueOrNull != null) ...[
                            const SizedBox(height: NxSpacing.s2),
                            _ClaimResultBanner(invite: claimState.valueOrNull!),
                          ],
                          const SizedBox(height: NxSpacing.s3),
                          NxButton(
                            label: 'Claim invite',
                            loading: claimState.isLoading,
                            onPressed: () => ref
                                .read(referralClaimProvider.notifier)
                                .claim(_claimController.text),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
              ),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, 0, NxSpacing.s4, NxSpacing.s2),
                  child: Text('Referral history', style: NxTypography.headlineSm),
                ),
              ),
              invitesAsync.when(
                data: (invites) {
                  if (invites.isEmpty) {
                    return SliverFillRemaining(
                      hasScrollBody: false,
                      child: NxEmptyState(title: l10n.emptyTitle, body: l10n.emptyMessage),
                    );
                  }
                  return SliverList.separated(
                    itemCount: invites.length,
                    separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
                    itemBuilder: (context, index) => Padding(
                      padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
                      child: _InviteTile(invite: invites[index]),
                    ),
                  );
                },
                loading: () => const SliverToBoxAdapter(child: Center(child: CircularProgressIndicator())),
                error: (e, _) => SliverToBoxAdapter(
                  child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(referralInvitesProvider)),
                ),
              ),
              const SliverToBoxAdapter(child: SizedBox(height: NxSpacing.s6)),
            ],
          ),
          error: (e, _) => ErrorView(message: e.toString(), onRetry: () => ref.invalidate(referralInfoProvider)),
        ),
      ),
    );
  }
}

class _ClaimResultBanner extends StatelessWidget {
  const _ClaimResultBanner({required this.invite});

  final ReferralInvite invite;

  @override
  Widget build(BuildContext context) {
    if (invite.fraudFlag || invite.claimStatus == ReferralClaimStatus.flagged) {
      return NxBanner(
        title: 'Claim flagged',
        message: invite.rejectedReason?.isNotEmpty == true
            ? invite.rejectedReason!
            : 'This claim was flagged for review.',
        variant: NxBannerVariant.warning,
      );
    }
    if (invite.claimStatus == ReferralClaimStatus.rejected) {
      return NxBanner(
        title: 'Claim rejected',
        message: invite.rejectedReason?.isNotEmpty == true
            ? invite.rejectedReason!
            : 'This invite code could not be claimed.',
        variant: NxBannerVariant.danger,
      );
    }
    return NxBanner(
      title: 'Claim ${invite.claimStatus.name}',
      message: invite.rejectedReason ?? 'Status: ${invite.claimStatus.name}',
      variant: NxBannerVariant.info,
    );
  }
}

class _InviteTile extends StatelessWidget {
  const _InviteTile({required this.invite});

  final ReferralInvite invite;

  @override
  Widget build(BuildContext context) {
    final statusColor = switch (invite.claimStatus) {
      ReferralClaimStatus.approved => context.nxColors.success,
      ReferralClaimStatus.rejected || ReferralClaimStatus.flagged => context.nxColors.danger,
      ReferralClaimStatus.expired => context.nxColors.textTertiary,
      _ => context.nxColors.warning,
    };

    final subtitleParts = <String>[invite.claimStatus.name];
    if (invite.fraudFlag) subtitleParts.add('fraud flagged');
    if (invite.rejectedReason != null && invite.rejectedReason!.isNotEmpty) {
      subtitleParts.add(invite.rejectedReason!);
    }

    return NxCard(
      child: ListTile(
        title: Text(invite.refereeLabel),
        subtitle: Text(subtitleParts.join(' · ')),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            if (invite.rewardMinor > 0)
              Text(Money(minorUnits: invite.rewardMinor, currency: invite.currency).format()),
            if (invite.fraudFlag)
              Icon(Icons.warning_amber, color: context.nxColors.warning, size: 18),
            Container(
              width: 8,
              height: 8,
              decoration: BoxDecoration(color: statusColor, shape: BoxShape.circle),
            ),
          ],
        ),
      ),
    );
  }
}
