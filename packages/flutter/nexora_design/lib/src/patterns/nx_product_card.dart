import 'package:flutter/material.dart';

import '../components/nx_card.dart';
import '../components/nx_icon_button.dart';
import '../components/nx_skeleton.dart';
import '../theme/nx_theme.dart';
import '../tokens/opacity.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_discount_badge.dart';
import 'nx_price_block.dart';
import 'nx_qty_selector.dart';
import 'nx_stock_indicator.dart';

enum NxProductCardVariant { grid, compact, horizontal }

/// Product card — grid, compact, or horizontal layout.
class NxProductCard extends StatelessWidget {
  const NxProductCard({
    super.key,
    required this.title,
    required this.price,
    this.variant = NxProductCardVariant.grid,
    this.imageUrl,
    this.unitMeta,
    this.originalPrice,
    this.discountLabel,
    this.stockStatus = NxStockStatus.inStock,
    this.lowStockLabel,
    this.quantity = 0,
    this.onTap,
    this.onAdd,
    this.onIncrement,
    this.onDecrement,
    this.onFavorite,
    this.isFavorite = false,
    this.semanticLabel,
  });

  final String title;
  final String price;
  final NxProductCardVariant variant;
  final String? imageUrl;
  final String? unitMeta;
  final String? originalPrice;
  final String? discountLabel;
  final NxStockStatus stockStatus;
  final String? lowStockLabel;
  final int quantity;
  final VoidCallback? onTap;
  final VoidCallback? onAdd;
  final VoidCallback? onIncrement;
  final VoidCallback? onDecrement;
  final VoidCallback? onFavorite;
  final bool isFavorite;
  final String? semanticLabel;

  bool get _outOfStock => stockStatus == NxStockStatus.outOfStock;

  @override
  Widget build(BuildContext context) {
    return switch (variant) {
      NxProductCardVariant.grid => _buildGrid(context),
      NxProductCardVariant.compact => _buildCompact(context),
      NxProductCardVariant.horizontal => _buildHorizontal(context),
    };
  }

  Widget _buildGrid(BuildContext context) {
    final colors = context.nxColors;

    return NxCard(
      variant: NxCardVariant.outlined,
      onTap: onTap,
      semanticLabel: semanticLabel ?? title,
      padding: EdgeInsets.zero,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          AspectRatio(
            aspectRatio: 1,
            child: Stack(
              children: [
                Positioned.fill(child: _image(context, radius: NxRadius.md)),
                if (discountLabel != null)
                  Positioned(
                    top: NxSpacing.s2,
                    left: NxSpacing.s2,
                    child: NxDiscountBadge(label: discountLabel!),
                  ),
                if (onFavorite != null)
                  Positioned(
                    top: NxSpacing.s1,
                    right: NxSpacing.s1,
                    child: NxIconButton(
                      icon: Icon(isFavorite ? Icons.favorite : Icons.favorite_border),
                      variant: NxIconButtonVariant.standard,
                      size: NxIconButtonSize.sm,
                      onPressed: onFavorite,
                      semanticLabel: 'Favorite',
                    ),
                  ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(NxSpacing.s3),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                if (unitMeta != null) ...[
                  const SizedBox(height: NxSpacing.s1),
                  Text(
                    unitMeta!,
                    style: NxTypography.captionMd.copyWith(color: colors.textTertiary),
                  ),
                ],
                const SizedBox(height: NxSpacing.s2),
                NxPriceBlock(price: price, originalPrice: originalPrice),
                const SizedBox(height: NxSpacing.s2),
                NxStockIndicator(status: stockStatus, lowStockLabel: lowStockLabel),
                const SizedBox(height: NxSpacing.s2),
                _actionRow(context),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCompact(BuildContext context) {
    final colors = context.nxColors;

    return NxCard(
      variant: NxCardVariant.outlined,
      onTap: onTap,
      semanticLabel: semanticLabel ?? title,
      child: Row(
        children: [
          SizedBox(
            width: 64,
            height: 64,
            child: _image(context, radius: NxRadius.sm),
          ),
          const SizedBox(width: NxSpacing.s3),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: NxTypography.titleSm.copyWith(color: colors.textPrimary),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: NxSpacing.s1),
                NxPriceBlock(
                  price: price,
                  originalPrice: originalPrice,
                  size: NxPriceBlockSize.sm,
                ),
              ],
            ),
          ),
          _actionRow(context, compact: true),
        ],
      ),
    );
  }

  Widget _buildHorizontal(BuildContext context) {
    final colors = context.nxColors;

    return NxCard(
      variant: NxCardVariant.outlined,
      onTap: onTap,
      semanticLabel: semanticLabel ?? title,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 96,
            height: 96,
            child: _image(context, radius: NxRadius.sm),
          ),
          const SizedBox(width: NxSpacing.s3),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                if (unitMeta != null) ...[
                  const SizedBox(height: NxSpacing.s1),
                  Text(
                    unitMeta!,
                    style: NxTypography.captionMd.copyWith(color: colors.textTertiary),
                  ),
                ],
                const SizedBox(height: NxSpacing.s2),
                NxPriceBlock(price: price, originalPrice: originalPrice),
                const SizedBox(height: NxSpacing.s2),
                NxStockIndicator(status: stockStatus, lowStockLabel: lowStockLabel),
                const SizedBox(height: NxSpacing.s2),
                _actionRow(context),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _image(BuildContext context, {required double radius}) {
    final colors = context.nxColors;

    Widget image;
    if (imageUrl != null && imageUrl!.isNotEmpty) {
      image = ClipRRect(
        borderRadius: BorderRadius.circular(radius),
        child: Image.network(
          imageUrl!,
          fit: BoxFit.cover,
          width: double.infinity,
          height: double.infinity,
          errorBuilder: (_, __, ___) => const NxSkeleton(borderRadius: null),
          loadingBuilder: (_, child, progress) {
            if (progress == null) return child;
            return NxSkeleton(
              borderRadius: BorderRadius.circular(radius),
              height: double.infinity,
            );
          },
        ),
      );
    } else {
      image = NxSkeleton(
        borderRadius: BorderRadius.circular(radius),
        height: double.infinity,
      );
    }

    if (_outOfStock) {
      image = ColorFiltered(
        colorFilter: ColorFilter.mode(
          colors.bgCanvas.withValues(alpha: NxOpacity.imageOutOfStock),
          BlendMode.saturation,
        ),
        child: Opacity(opacity: 1 - NxOpacity.imageOutOfStock, child: image),
      );
    }

    return image;
  }

  Widget _actionRow(BuildContext context, {bool compact = false}) {
    if (_outOfStock) {
      return Align(
        alignment: Alignment.centerRight,
        child: Text(
          'Unavailable',
          style: NxTypography.captionMd.copyWith(
            color: context.nxColors.textDisabled,
          ),
        ),
      );
    }

    if (quantity > 0 && onIncrement != null && onDecrement != null) {
      return NxQtySelector(
        quantity: quantity,
        onIncrement: onIncrement!,
        onDecrement: onDecrement!,
        size: compact ? NxQtySelectorSize.sm : NxQtySelectorSize.md,
      );
    }

    return Align(
      alignment: Alignment.centerRight,
      child: NxIconButton(
        icon: const Icon(Icons.add),
        variant: NxIconButtonVariant.filled,
        size: compact ? NxIconButtonSize.sm : NxIconButtonSize.md,
        onPressed: onAdd,
        semanticLabel: 'Add to cart',
      ),
    );
  }
}
