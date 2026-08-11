import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../shared/utils/money.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/offer_entity.dart';
import '../providers/offers_providers.dart';

class OffersScreen extends ConsumerWidget {
  const OffersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    ref.watch(offersRealtimeInvalidationProvider);
    final offersAsync = ref.watch(offersProvider);

    return Scaffold(
      appBar: const NxTopBar(title: 'Offers'),
      body: AsyncValueWidget<List<Offer>>(
        value: offersAsync,
        data: (offers) {
          if (offers.isEmpty) {
            return const NxEmptyState(
              title: 'Nothing here yet',
              body: 'Check back soon.',
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(offersProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: offers.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) => _OfferTile(offer: offers[index]),
            ),
          );
        },
      ),
    );
  }
}

class _OfferTile extends ConsumerStatefulWidget {
  const _OfferTile({required this.offer});
  final Offer offer;

  @override
  ConsumerState<_OfferTile> createState() => _OfferTileState();
}

class _OfferTileState extends ConsumerState<_OfferTile> {
  bool _busy = false;

  Future<void> _act(Future<dynamic> Function() action) async {
    setState(() => _busy = true);
    await action();
    if (mounted) {
      setState(() => _busy = false);
      ref.invalidate(offersProvider);
    }
  }

  @override
  Widget build(BuildContext context) {
    final offer = widget.offer;
    final colors = context.nxColors;
    final payout =
        Money(minorUnits: offer.payoutMinor, currency: offer.currency);

    return NxCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  offer.storeName,
                  style:
                      NxTypography.titleSm.copyWith(color: colors.textPrimary),
                ),
              ),
              Text(
                payout.format(),
                style: NxTypography.priceMd.copyWith(color: colors.textBrand),
              ),
            ],
          ),
          const SizedBox(height: NxSpacing.s1),
          Text(
            offer.customerArea,
            style: NxTypography.bodySm.copyWith(color: colors.textSecondary),
          ),
          if (offer.batchId != null)
            Text(
              'Batch ${offer.batchId}',
              style:
                  NxTypography.captionMd.copyWith(color: colors.textTertiary),
            ),
          const SizedBox(height: NxSpacing.s3),
          Row(
            children: [
              Expanded(
                child: NxButton(
                  label: 'Reject',
                  variant: NxButtonVariant.secondary,
                  loading: _busy,
                  onPressed: _busy
                      ? null
                      : () => _act(
                            () =>
                                ref.read(offerActionsProvider).reject(offer.id),
                          ),
                ),
              ),
              const SizedBox(width: NxSpacing.s2),
              Expanded(
                child: NxButton(
                  label: 'Accept',
                  loading: _busy,
                  onPressed: _busy || offer.isExpired
                      ? null
                      : () => _act(
                            () =>
                                ref.read(offerActionsProvider).accept(offer.id),
                          ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
