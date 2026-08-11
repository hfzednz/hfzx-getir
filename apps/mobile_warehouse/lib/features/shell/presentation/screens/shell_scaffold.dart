import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

/// Dense warehouse bottom nav: Home, Picking, Packing, Dispatch, More.
class ShellScaffold extends ConsumerWidget {
  const ShellScaffold({super.key, required this.navigationShell});

  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final colors = context.nxColors;
    final brightness = Theme.of(context).brightness;

    final items = <(IconData, String)>[
      (Icons.dashboard_outlined, 'Home'),
      (Icons.shopping_basket_outlined, 'Pick'),
      (Icons.inventory_2_outlined, 'Pack'),
      (Icons.local_shipping_outlined, 'Dispatch'),
      (Icons.more_horiz, 'More'),
    ];

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: DecoratedBox(
        decoration: BoxDecoration(
          color: colors.bgNav,
          boxShadow: NxElevation.forBrightness(brightness, 1),
          border: Border(top: BorderSide(color: colors.borderSubtle)),
        ),
        child: SafeArea(
          top: false,
          child: SizedBox(
            height: 52,
            child: Row(
              children: List.generate(items.length, (index) {
                final (icon, label) = items[index];
                final selected = navigationShell.currentIndex == index;
                return Expanded(
                  child: Semantics(
                    button: true,
                    selected: selected,
                    label: label,
                    child: InkWell(
                      onTap: () => navigationShell.goBranch(
                        index,
                        initialLocation:
                            index == navigationShell.currentIndex,
                      ),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            icon,
                            size: NxIconSize.md,
                            color: selected
                                ? colors.navItemActive
                                : colors.navItemDefault,
                          ),
                          const SizedBox(height: NxSpacing.s0_5),
                          Text(
                            label,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: NxTypography.navMd.copyWith(
                              color: selected
                                  ? colors.navItemActive
                                  : colors.navItemDefault,
                              fontSize: 10,
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
      ),
    );
  }
}
