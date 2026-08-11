import 'package:flutter/material.dart';

import '../tokens/colors.dart';
import '../tokens/opacity.dart';
import '../tokens/radius.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// Surface density profile for spacing adjustments.
enum NxDensity {
  comfortable,
  compact,
  dense,
}

/// Theme extension holding NEXORA semantic tokens.
@immutable
class NxThemeExtension extends ThemeExtension<NxThemeExtension> {
  const NxThemeExtension({
    required this.colors,
    required this.density,
    required this.brightness,
  });

  final NxColorRoles colors;
  final NxDensity density;
  final Brightness brightness;

  double get verticalPaddingMultiplier => switch (density) {
        NxDensity.comfortable => 1.0,
        NxDensity.compact => 0.85,
        NxDensity.dense => 0.75,
      };

  double scaledSpacing(double base) => base * verticalPaddingMultiplier;

  @override
  NxThemeExtension copyWith({
    NxColorRoles? colors,
    NxDensity? density,
    Brightness? brightness,
  }) {
    return NxThemeExtension(
      colors: colors ?? this.colors,
      density: density ?? this.density,
      brightness: brightness ?? this.brightness,
    );
  }

  @override
  NxThemeExtension lerp(covariant ThemeExtension<NxThemeExtension>? other, double t) {
    if (other is! NxThemeExtension) return this;
    return NxThemeExtension(
      colors: colors.lerp(other.colors, t),
      density: t < 0.5 ? density : other.density,
      brightness: t < 0.5 ? brightness : other.brightness,
    );
  }
}

/// NEXORA theme factory.
abstract final class NxTheme {
  static ThemeData light({NxDensity density = NxDensity.comfortable}) =>
      _buildTheme(
        brightness: Brightness.light,
        colors: NxColorRoles.light,
        density: density,
      );

  static ThemeData dark({NxDensity density = NxDensity.comfortable}) =>
      _buildTheme(
        brightness: Brightness.dark,
        colors: NxColorRoles.dark,
        density: density,
      );

  static ThemeData _buildTheme({
    required Brightness brightness,
    required NxColorRoles colors,
    required NxDensity density,
  }) {
    final extension = NxThemeExtension(
      colors: colors,
      density: density,
      brightness: brightness,
    );

    final colorScheme = ColorScheme(
      brightness: brightness,
      primary: colors.bgBrand,
      onPrimary: colors.textOnBrand,
      secondary: colors.bgAccent,
      onSecondary: colors.textOnAccent,
      error: colors.danger,
      onError: colors.textOnBrand,
      surface: colors.bgSurface,
      onSurface: colors.textPrimary,
    );

    return ThemeData(
      useMaterial3: true,
      brightness: brightness,
      colorScheme: colorScheme,
      scaffoldBackgroundColor: colors.bgCanvas,
      canvasColor: colors.bgCanvas,
      dividerColor: colors.borderSubtle,
      disabledColor: colors.textDisabled,
      splashColor: colors.bgBrand.withValues(alpha: NxOpacity.hover),
      highlightColor: colors.bgBrand.withValues(alpha: NxOpacity.pressed),
      fontFamily: NxFontFamily.body,
      fontFamilyFallback: NxFontFamily.bodyFallback,
      textTheme: NxTypography.textTheme(color: colors.textPrimary),
      extensions: [extension],
      appBarTheme: AppBarTheme(
        elevation: 0,
        scrolledUnderElevation: 0,
        backgroundColor: colors.bgSurface,
        foregroundColor: colors.textPrimary,
        titleTextStyle: NxTypography.headlineSm.copyWith(color: colors.textPrimary),
      ),
      snackBarTheme: SnackBarThemeData(
        backgroundColor: colors.bgSurfaceRaised,
        contentTextStyle: NxTypography.bodyMd.copyWith(color: colors.textPrimary),
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          side: BorderSide(color: colors.borderSubtle),
        ),
      ),
      dialogTheme: DialogThemeData(
        backgroundColor: colors.bgSurface,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(NxRadius.lg),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: colors.bgSurface,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: NxSpacing.s4,
          vertical: NxSpacing.s3,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          borderSide: BorderSide(color: colors.borderDefault),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          borderSide: BorderSide(color: colors.borderDefault),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          borderSide: BorderSide(color: colors.borderFocus, width: NxBorderWidth.thick),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(NxRadius.md),
          borderSide: BorderSide(color: colors.borderDanger, width: NxBorderWidth.thick),
        ),
        labelStyle: NxTypography.bodySm.copyWith(color: colors.textSecondary),
        hintStyle: NxTypography.bodyMd.copyWith(color: colors.textTertiary),
        errorStyle: NxTypography.captionMd.copyWith(color: colors.danger),
      ),
    );
  }
}

/// Convenience accessor for [NxThemeExtension].
extension NxThemeContext on BuildContext {
  NxThemeExtension get nx => Theme.of(this).extension<NxThemeExtension>()!;
  NxColorRoles get nxColors => nx.colors;
}
