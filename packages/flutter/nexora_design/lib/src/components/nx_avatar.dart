import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/colors.dart';
import '../tokens/radius.dart';
import '../tokens/typography.dart';

enum NxAvatarSize { sm, md, lg, xl }

/// Avatar — image, initials, or icon fallback.
class NxAvatar extends StatelessWidget {
  const NxAvatar({
    super.key,
    this.imageUrl,
    this.initials,
    this.icon,
    this.size = NxAvatarSize.md,
    this.showPresence = false,
    this.presenceOnline = false,
    this.semanticLabel,
  });

  final String? imageUrl;
  final String? initials;
  final IconData? icon;
  final NxAvatarSize size;
  final bool showPresence;
  final bool presenceOnline;
  final String? semanticLabel;

  double get _dimension => switch (size) {
        NxAvatarSize.sm => 24,
        NxAvatarSize.md => 32,
        NxAvatarSize.lg => 40,
        NxAvatarSize.xl => 56,
      };

  TextStyle get _initialStyle => switch (size) {
        NxAvatarSize.sm => NxTypography.captionSm,
        NxAvatarSize.md => NxTypography.captionMd,
        NxAvatarSize.lg => NxTypography.bodySm,
        NxAvatarSize.xl => NxTypography.titleSm,
      };

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final dimension = _dimension;

    Widget content;
    if (imageUrl != null && imageUrl!.isNotEmpty) {
      content = ClipRRect(
        borderRadius: BorderRadius.circular(NxRadius.full),
        child: Image.network(
          imageUrl!,
          width: dimension,
          height: dimension,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => _fallback(colors, dimension),
        ),
      );
    } else {
      content = _fallback(colors, dimension);
    }

    if (!showPresence) {
      return Semantics(label: semanticLabel, child: content);
    }

    return Semantics(
      label: semanticLabel,
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          content,
          Positioned(
            right: 0,
            bottom: 0,
            child: Container(
              width: dimension * 0.28,
              height: dimension * 0.28,
              decoration: BoxDecoration(
                color: presenceOnline ? colors.success : colors.textTertiary,
                shape: BoxShape.circle,
                border: Border.all(color: colors.bgSurface, width: 2),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _fallback(NxColorRoles colors, double dimension) {
    return Container(
      width: dimension,
      height: dimension,
      decoration: BoxDecoration(
        color: colors.bgSunken,
        shape: BoxShape.circle,
        border: Border.all(color: colors.borderSubtle),
      ),
      alignment: Alignment.center,
      child: initials != null && initials!.isNotEmpty
          ? Text(
              initials!.substring(0, initials!.length.clamp(0, 2)).toUpperCase(),
              style: _initialStyle.copyWith(color: colors.textPrimary),
            )
          : Icon(
              icon ?? Icons.person,
              size: dimension * 0.5,
              color: colors.iconSecondary,
            ),
    );
  }
}
