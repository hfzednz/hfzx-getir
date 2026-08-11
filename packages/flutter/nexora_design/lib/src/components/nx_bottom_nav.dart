import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/elevation.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';
import 'nx_badge.dart';

enum NxBottomNavItem { home, categories, search, cart, account }

/// Customer bottom navigation — 5 fixed destinations.
class NxBottomNav extends StatelessWidget {
  const NxBottomNav({
    super.key,
    required this.currentIndex,
    required this.onTap,
    this.cartCount = 0,
  });

  final int currentIndex;
  final ValueChanged<int> onTap;
  final int cartCount;

  static const List<NxBottomNavItem> items = [
    NxBottomNavItem.home,
    NxBottomNavItem.categories,
    NxBottomNavItem.search,
    NxBottomNavItem.cart,
    NxBottomNavItem.account,
  ];

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;

    return DecoratedBox(
      decoration: BoxDecoration(
        color: colors.bgNav,
        boxShadow: NxElevation.forBrightness(brightness, 1),
        border: Border(top: BorderSide(color: colors.borderSubtle)),
      ),
      child: SafeArea(
        top: false,
        child: SizedBox(
          height: 64,
          child: Row(
            children: List.generate(items.length, (index) {
              final item = items[index];
              final selected = currentIndex == index;
              final (icon, label) = _itemMeta(item);

              return Expanded(
                child: Semantics(
                  button: true,
                  selected: selected,
                  label: label,
                  child: InkWell(
                    onTap: () => onTap(index),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Stack(
                          clipBehavior: Clip.none,
                          children: [
                            Icon(
                              icon,
                              size: NxIconSize.lg,
                              color: selected ? colors.navItemActive : colors.navItemDefault,
                            ),
                            if (item == NxBottomNavItem.cart && cartCount > 0)
                              Positioned(
                                right: -8,
                                top: -4,
                                child: NxBadge(count: cartCount),
                              ),
                          ],
                        ),
                        const SizedBox(height: NxSpacing.s1),
                        Text(
                          label,
                          style: NxTypography.navMd.copyWith(
                            color: selected ? colors.navItemActive : colors.navItemDefault,
                          ),
                        ),
                        if (selected)
                          Container(
                            margin: const EdgeInsets.only(top: NxSpacing.s1),
                            width: 20,
                            height: 2,
                            decoration: BoxDecoration(
                              color: colors.navIndicator,
                              borderRadius: BorderRadius.circular(NxRadius.full),
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
              );
            }),
          ),
        ),
      ),
    );
  }

  (IconData, String) _itemMeta(NxBottomNavItem item) => switch (item) {
        NxBottomNavItem.home => (Icons.home_outlined, 'Home'),
        NxBottomNavItem.categories => (Icons.grid_view_outlined, 'Categories'),
        NxBottomNavItem.search => (Icons.search, 'Search'),
        NxBottomNavItem.cart => (Icons.shopping_cart_outlined, 'Cart'),
        NxBottomNavItem.account => (Icons.person_outline, 'Account'),
      };
}
