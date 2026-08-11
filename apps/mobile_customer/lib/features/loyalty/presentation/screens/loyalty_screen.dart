import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/loyalty_entity.dart';
import '../providers/loyalty_providers.dart';

class LoyaltyScreen extends ConsumerWidget {
  const LoyaltyScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final accountAsync = ref.watch(loyaltyAccountProvider);
    final achievementsAsync = ref.watch(loyaltyAchievementsProvider);
    final milestonesAsync = ref.watch(loyaltyMilestonesProvider);
    final badgesAsync = ref.watch(loyaltyBadgesProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.loyaltyTitle),
      body: RefreshIndicator(
        onRefresh: () async {
          ref.invalidate(loyaltyAccountProvider);
          ref.invalidate(loyaltyAchievementsProvider);
          ref.invalidate(loyaltyMilestonesProvider);
          ref.invalidate(loyaltyBadgesProvider);
        },
        child: AsyncValueWidget(
          value: accountAsync,
          data: (account) => CustomScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            slivers: [
              SliverToBoxAdapter(child: _TierProgressCard(account: account, ref: ref)),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, NxSpacing.s4, NxSpacing.s4, NxSpacing.s2),
                  child: Text('Milestones', style: NxTypography.headlineSm),
                ),
              ),
              _buildMilestones(milestonesAsync, ref),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, NxSpacing.s4, NxSpacing.s4, NxSpacing.s2),
                  child: Text('Achievements', style: NxTypography.headlineSm),
                ),
              ),
              _buildAchievements(achievementsAsync, ref),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(NxSpacing.s4, NxSpacing.s4, NxSpacing.s4, NxSpacing.s2),
                  child: Text('Badges', style: NxTypography.headlineSm),
                ),
              ),
              _buildBadges(badgesAsync, ref),
              const SliverToBoxAdapter(child: SizedBox(height: NxSpacing.s6)),
            ],
          ),
          error: (e, _) => ErrorView(message: e.toString(), onRetry: () => ref.invalidate(loyaltyAccountProvider)),
        ),
      ),
    );
  }

  Widget _buildMilestones(AsyncValue<List<LoyaltyMilestone>> async, WidgetRef ref) {
    return async.when(
      data: (items) => SliverList.separated(
        itemCount: items.length,
        separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
        itemBuilder: (context, index) {
          final m = items[index];
          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
            child: NxCard(
              child: ListTile(
                leading: Icon(m.reached ? Icons.flag : Icons.flag_outlined, color: context.nxColors.iconBrand),
                title: Text(m.title),
                subtitle: Text('${m.targetPoints} pts · ${m.rewardLabel}'),
                trailing: m.reached ? Icon(Icons.check_circle, color: context.nxColors.success) : null,
              ),
            ),
          );
        },
      ),
      loading: () => const SliverToBoxAdapter(child: Center(child: NxSpinner())),
      error: (e, _) => SliverToBoxAdapter(child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(loyaltyMilestonesProvider))),
    );
  }

  Widget _buildAchievements(AsyncValue<List<LoyaltyAchievement>> async, WidgetRef ref) {
    return async.when(
      data: (items) => SliverList.separated(
        itemCount: items.length,
        separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
        itemBuilder: (context, index) {
          final a = items[index];
          return Padding(
            padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
            child: NxCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  ListTile(
                    leading: Icon(a.unlocked ? Icons.emoji_events : Icons.lock_outline),
                    title: Text(a.title),
                    subtitle: Text(a.description),
                  ),
                  if (!a.unlocked)
                    Padding(
                      padding: const EdgeInsets.fromLTRB(NxSpacing.s4, 0, NxSpacing.s4, NxSpacing.s3),
                      child: LinearProgressIndicator(value: a.progressPercent / 100),
                    ),
                ],
              ),
            ),
          );
        },
      ),
      loading: () => const SliverToBoxAdapter(child: Center(child: NxSpinner())),
      error: (e, _) => SliverToBoxAdapter(child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(loyaltyAchievementsProvider))),
    );
  }

  Widget _buildBadges(AsyncValue<List<LoyaltyBadge>> async, WidgetRef ref) {
    return async.when(
      data: (items) {
        if (items.isEmpty) {
          return const SliverToBoxAdapter(child: SizedBox.shrink());
        }
        return SliverToBoxAdapter(
          child: SizedBox(
            height: 100,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              padding: const EdgeInsets.symmetric(horizontal: NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(width: NxSpacing.s2),
              itemBuilder: (context, index) {
                final b = items[index];
                return NxChip(label: b.label);
              },
            ),
          ),
        );
      },
      loading: () => const SliverToBoxAdapter(child: Center(child: NxSpinner())),
      error: (e, _) => SliverToBoxAdapter(child: ErrorView(message: e.toString(), onRetry: () => ref.invalidate(loyaltyBadgesProvider))),
    );
  }
}

class _TierProgressCard extends StatelessWidget {
  const _TierProgressCard({required this.account, required this.ref});

  final LoyaltyAccount account;
  final WidgetRef ref;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(NxSpacing.s4),
      child: NxCard(
        child: Padding(
          padding: const EdgeInsets.all(NxSpacing.s4),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(account.tierLabel, style: NxTypography.displayLg),
                  Text('${account.points} pts', style: NxTypography.headlineSm),
                ],
              ),
              const SizedBox(height: NxSpacing.s3),
              LinearProgressIndicator(value: (account.tierProgressPercent / 100).clamp(0.0, 1.0)),
              const SizedBox(height: NxSpacing.s1),
              Text(
                account.nextTierPoints > 0
                    ? '${account.nextTierPoints - account.points} pts to next tier'
                    : 'Top tier reached',
                style: NxTypography.captionMd,
              ),
              if (account.dailyRewardAvailable && !account.dailyRewardClaimedToday) ...[
                const SizedBox(height: NxSpacing.s4),
                NxButton(
                  label: 'Claim daily reward',
                  onPressed: () => ref.read(dailyRewardClaimProvider.notifier).claim(),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
