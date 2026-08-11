import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/typography.dart';

/// Top app bar with title, optional subtitle, and trailing actions.
class NxTopBar extends StatelessWidget implements PreferredSizeWidget {
  const NxTopBar({
    super.key,
    this.title,
    this.subtitle,
    this.leading,
    this.actions = const [],
    this.centerTitle = false,
    this.semanticLabel,
  });

  final String? title;
  final String? subtitle;
  final Widget? leading;
  final List<Widget> actions;
  final bool centerTitle;
  final String? semanticLabel;

  @override
  Size get preferredSize => Size.fromHeight(subtitle != null ? 72 : kToolbarHeight);

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;

    return AppBar(
      backgroundColor: colors.bgSurface,
      foregroundColor: colors.textPrimary,
      elevation: 0,
      scrolledUnderElevation: 0,
      centerTitle: centerTitle,
      leading: leading,
      actions: actions,
      title: Semantics(
        header: true,
        label: semanticLabel ?? title,
        child: subtitle != null
            ? Column(
                crossAxisAlignment:
                    centerTitle ? CrossAxisAlignment.center : CrossAxisAlignment.start,
                children: [
                  if (title != null)
                    Text(
                      title!,
                      style: NxTypography.titleMd.copyWith(color: colors.textPrimary),
                    ),
                  Text(
                    subtitle!,
                    style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
                  ),
                ],
              )
            : title != null
                ? Text(
                    title!,
                    style: NxTypography.headlineSm.copyWith(color: colors.textPrimary),
                  )
                : null,
      ),
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(1),
        child: Divider(height: 1, color: colors.borderSubtle),
      ),
    );
  }
}
