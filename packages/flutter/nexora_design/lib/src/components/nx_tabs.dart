import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';
import '../tokens/typography.dart';

/// Scrollable tabs with brand underline indicator.
class NxTabs extends StatefulWidget {
  const NxTabs({
    super.key,
    required this.tabs,
    required this.children,
    this.initialIndex = 0,
    this.onChanged,
  });

  final List<String> tabs;
  final List<Widget> children;
  final int initialIndex;
  final ValueChanged<int>? onChanged;

  @override
  State<NxTabs> createState() => _NxTabsState();
}

class _NxTabsState extends State<NxTabs> with SingleTickerProviderStateMixin {
  late TabController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TabController(
      length: widget.tabs.length,
      vsync: this,
      initialIndex: widget.initialIndex,
    )..addListener(() {
        if (!_controller.indexIsChanging) {
          widget.onChanged?.call(_controller.index);
        }
      });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final equalWidth = widget.tabs.length <= 4;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Material(
          color: colors.bgSurface,
          child: TabBar(
            controller: _controller,
            isScrollable: !equalWidth,
            tabAlignment: equalWidth ? TabAlignment.fill : TabAlignment.start,
            indicatorColor: colors.navIndicator,
            indicatorWeight: 2,
            labelColor: colors.navItemActive,
            unselectedLabelColor: colors.navItemDefault,
            labelStyle: NxTypography.titleSm,
            unselectedLabelStyle: NxTypography.bodyMd,
            dividerColor: colors.borderSubtle,
            tabs: widget.tabs.map((t) => Tab(text: t)).toList(),
          ),
        ),
        Expanded(
          child: TabBarView(
            controller: _controller,
            children: widget.children,
          ),
        ),
      ],
    );
  }
}
