import 'package:flutter/material.dart';

import 'colors.dart';

/// Elevation shadow tokens mapped to Flutter [BoxShadow] lists.
abstract final class NxElevation {
  static const List<BoxShadow> level0 = [];

  static const List<BoxShadow> level1 = [
    BoxShadow(
      color: Color(0x0F0B1214),
      offset: Offset(0, 1),
      blurRadius: 2,
    ),
    BoxShadow(
      color: Color(0x0A0B1214),
      offset: Offset(0, 1),
      blurRadius: 1,
    ),
  ];

  static const List<BoxShadow> level2 = [
    BoxShadow(
      color: Color(0x140B1214),
      offset: Offset(0, 4),
      blurRadius: 12,
    ),
  ];

  static const List<BoxShadow> level3 = [
    BoxShadow(
      color: Color(0x1F0B1214),
      offset: Offset(0, 8),
      blurRadius: 24,
    ),
  ];

  static const List<BoxShadow> level4 = [
    BoxShadow(
      color: Color(0x290B1214),
      offset: Offset(0, 16),
      blurRadius: 40,
    ),
  ];

  /// Dark mode prefers surface shifts over shadows.
  static List<BoxShadow> forBrightness(Brightness brightness, int level) {
    if (brightness == Brightness.dark && level <= 1) {
      return level0;
    }
    return switch (level) {
      0 => level0,
      1 => level1,
      2 => level2,
      3 => level3,
      _ => level4,
    };
  }

  static BoxDecoration decoration({
    required Brightness brightness,
    required int level,
    Color? color,
    BorderRadius? borderRadius,
    Border? border,
  }) {
    return BoxDecoration(
      color: color ?? NxNeutralColors.n0,
      borderRadius: borderRadius,
      border: border,
      boxShadow: forBrightness(brightness, level),
    );
  }
}
