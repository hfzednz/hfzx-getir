import 'package:flutter/material.dart';

/// Font family tokens with system fallbacks.
abstract final class NxFontFamily {
  static const String display = 'Satoshi';
  static const String body = 'Geist';
  static const String mono = 'Geist Mono';

  static const List<String> displayFallback = [
    'ui-rounded',
    'system-ui',
    'Segoe UI',
    'Roboto',
    'sans-serif',
  ];

  static const List<String> bodyFallback = [
    'system-ui',
    'Segoe UI',
    'Roboto',
    'sans-serif',
  ];

  static const List<String> monoFallback = [
    'ui-monospace',
    'Cascadia Mono',
    'Consolas',
    'monospace',
  ];
}

/// Typography scale from [nexora.tokens.json].
abstract final class NxTypography {
  static TextStyle _base({
    required String family,
    required List<String> fallback,
    required double size,
    required double lineHeight,
    required FontWeight weight,
    required double tracking,
    bool tabular = false,
  }) {
    final height = lineHeight / size;
    return TextStyle(
      fontFamily: family,
      fontFamilyFallback: fallback,
      fontSize: size,
      height: height,
      fontWeight: weight,
      letterSpacing: size * tracking,
      fontFeatures: tabular ? const [FontFeature.tabularFigures()] : null,
    );
  }

  static TextStyle get displayXl => _base(
        family: NxFontFamily.display,
        fallback: NxFontFamily.displayFallback,
        size: 40,
        lineHeight: 48,
        weight: FontWeight.w700,
        tracking: -0.02,
      );

  static TextStyle get displayLg => _base(
        family: NxFontFamily.display,
        fallback: NxFontFamily.displayFallback,
        size: 32,
        lineHeight: 40,
        weight: FontWeight.w700,
        tracking: -0.02,
      );

  static TextStyle get headlineLg => _base(
        family: NxFontFamily.display,
        fallback: NxFontFamily.displayFallback,
        size: 28,
        lineHeight: 36,
        weight: FontWeight.w700,
        tracking: -0.015,
      );

  static TextStyle get headlineMd => _base(
        family: NxFontFamily.display,
        fallback: NxFontFamily.displayFallback,
        size: 24,
        lineHeight: 32,
        weight: FontWeight.w600,
        tracking: -0.01,
      );

  static TextStyle get headlineSm => _base(
        family: NxFontFamily.display,
        fallback: NxFontFamily.displayFallback,
        size: 20,
        lineHeight: 28,
        weight: FontWeight.w600,
        tracking: -0.01,
      );

  static TextStyle get titleLg => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 18,
        lineHeight: 26,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get titleMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 16,
        lineHeight: 24,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get titleSm => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 14,
        lineHeight: 20,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get bodyLg => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 16,
        lineHeight: 24,
        weight: FontWeight.w400,
        tracking: 0,
      );

  static TextStyle get bodyMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 14,
        lineHeight: 20,
        weight: FontWeight.w400,
        tracking: 0,
      );

  static TextStyle get bodySm => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 13,
        lineHeight: 18,
        weight: FontWeight.w400,
        tracking: 0,
      );

  static TextStyle get captionMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 12,
        lineHeight: 16,
        weight: FontWeight.w400,
        tracking: 0.01,
      );

  static TextStyle get captionSm => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 11,
        lineHeight: 14,
        weight: FontWeight.w500,
        tracking: 0.02,
      );

  static TextStyle get overline => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 11,
        lineHeight: 14,
        weight: FontWeight.w600,
        tracking: 0.08,
      );

  static TextStyle get buttonMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 14,
        lineHeight: 20,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get buttonSm => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 13,
        lineHeight: 18,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get buttonLg => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 16,
        lineHeight: 24,
        weight: FontWeight.w600,
        tracking: 0,
      );

  static TextStyle get navMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 12,
        lineHeight: 16,
        weight: FontWeight.w500,
        tracking: 0,
      );

  static TextStyle get priceMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 16,
        lineHeight: 24,
        weight: FontWeight.w700,
        tracking: 0,
        tabular: true,
      );

  static TextStyle get priceLg => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 20,
        lineHeight: 28,
        weight: FontWeight.w700,
        tracking: -0.01,
        tabular: true,
      );

  static TextStyle get priceSm => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 14,
        lineHeight: 20,
        weight: FontWeight.w600,
        tracking: 0,
        tabular: true,
      );

  static TextStyle get etaMd => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 16,
        lineHeight: 24,
        weight: FontWeight.w700,
        tracking: 0,
        tabular: true,
      );

  static TextStyle get tableCell => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 13,
        lineHeight: 18,
        weight: FontWeight.w400,
        tracking: 0,
        tabular: true,
      );

  static TextStyle get dashKpi => _base(
        family: NxFontFamily.body,
        fallback: NxFontFamily.bodyFallback,
        size: 32,
        lineHeight: 40,
        weight: FontWeight.w700,
        tracking: -0.01,
        tabular: true,
      );

  static TextTheme textTheme({Color? color}) {
    return TextTheme(
      displayLarge: displayXl.copyWith(color: color),
      displayMedium: displayLg.copyWith(color: color),
      headlineLarge: headlineLg.copyWith(color: color),
      headlineMedium: headlineMd.copyWith(color: color),
      headlineSmall: headlineSm.copyWith(color: color),
      titleLarge: titleLg.copyWith(color: color),
      titleMedium: titleMd.copyWith(color: color),
      titleSmall: titleSm.copyWith(color: color),
      bodyLarge: bodyLg.copyWith(color: color),
      bodyMedium: bodyMd.copyWith(color: color),
      bodySmall: bodySm.copyWith(color: color),
      labelLarge: buttonMd.copyWith(color: color),
      labelMedium: captionMd.copyWith(color: color),
      labelSmall: captionSm.copyWith(color: color),
    );
  }
}
