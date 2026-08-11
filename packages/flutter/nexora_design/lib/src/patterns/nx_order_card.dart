import 'package:flutter/material.dart';

import '../components/nx_card.dart';
import '../components/nx_skeleton.dart';
import '../theme/nx_theme.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_price_block.dart';

/// Order list card — id, status, ETA, total, item thumbs.
class NxOrderCard extends StatelessWidget {
  const NxOrderCard({
    super.key,
    required this.orderId,
    required this.statusLabel,
    required this.total,
    this.statusColor,
    this.subtitle,
    this.imageUrls = const [],
    this.extraItemCount = 0,
    this.onTap,
    this.semanticLabel,
  });

  final String orderId;
  final String statusLabel;
  final String total;
  final Color? statusColor;
  final String? subtitle;
  final List<String> imageUrls;
  final int extraItemCount;
  final VoidCallback? onTap;
  final String? semanticLabel;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final statusFg = statusColor ?? colors.info;

    return NxCard(
      variant: NxCardVariant.interactive,
      onTap: onTap,
      semanticLabel: semanticLabel ?? 'Order $orderId',
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Text(
                  orderId,
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: NxSpacing.s2,
                  vertical: NxSpacing.s1,
                ),
                decoration: BoxDecoration(
                  color: statusFg.withValues(alpha: 0.12),
                  borderRadius: BorderRadius.circular(NxRadius.xs),
                ),
                child: Text(
                  statusLabel,
                  style: NxTypography.captionSm.copyWith(color: statusFg),
                ),
              ),
            ],
          ),
          if (subtitle != null) ...[
            const SizedBox(height: NxSpacing.s1),
            Text(
              subtitle!,
              style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
            ),
          ],
          const SizedBox(height: NxSpacing.s3),
          Row(
            children: [
              ...imageUrls.take(4).map(_thumb),
              if (extraItemCount > 0)
                Container(
                  width: 40,
                  height: 40,
                  margin: const EdgeInsets.only(right: NxSpacing.s2),
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: colors.bgSunken,
                    borderRadius: BorderRadius.circular(NxRadius.sm),
                  ),
                  child: Text(
                    '+$extraItemCount',
                    style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
                  ),
                ),
              const Spacer(),
              NxPriceBlock(price: total, size: NxPriceBlockSize.sm),
            ],
          ),
        ],
      ),
    );
  }

  Widget _thumb(String url) {
    return Padding(
      padding: const EdgeInsets.only(right: NxSpacing.s2),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(NxRadius.sm),
        child: Image.network(
          url,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => const NxSkeleton(width: 40, height: 40),
        ),
      ),
    );
  }
}
