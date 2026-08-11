import 'package:flutter/material.dart';

/// Responsive breakpoint tokens from [nexora.tokens.json].
abstract final class NxBreakpoints {
  static const double xs = 0;
  static const double sm = 600;
  static const double md = 905;
  static const double lg = 1240;
  static const double xl = 1440;
  static const double xxl = 1920;
}

/// Breakpoint helpers for adaptive layouts.
abstract final class NxBreakpointUtils {
  static double widthOf(BuildContext context) =>
      MediaQuery.sizeOf(context).width;

  static bool isMobile(BuildContext context) => widthOf(context) < NxBreakpoints.sm;

  static bool isTablet(BuildContext context) {
    final w = widthOf(context);
    return w >= NxBreakpoints.sm && w < NxBreakpoints.lg;
  }

  static bool isDesktop(BuildContext context) => widthOf(context) >= NxBreakpoints.lg;

  static T valueFor<T>({
    required BuildContext context,
    required T mobile,
    T? tablet,
    T? desktop,
  }) {
    if (isDesktop(context)) return desktop ?? tablet ?? mobile;
    if (isTablet(context)) return tablet ?? mobile;
    return mobile;
  }
}
