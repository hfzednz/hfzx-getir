import 'package:flutter/material.dart';

/// Brand color primitives from [nexora.tokens.json].
abstract final class NxBrandColors {
  static const Color primary50 = Color(0xFFEBF8F8);
  static const Color primary100 = Color(0xFFD7F0F0);
  static const Color primary500 = Color(0xFF0F8585);
  static const Color primary600 = Color(0xFF0B6E6E);
  static const Color primary700 = Color(0xFF085858);
  static const Color primary800 = Color(0xFF064545);

  static const Color ink = Color(0xFF0B1214);

  static const Color accent100 = Color(0xFFF7FAD4);
  static const Color accent400 = Color(0xFFE8F07A);
  static const Color accent500 = Color(0xFFD4DC5C);
}

/// Neutral (graphite) scale.
abstract final class NxNeutralColors {
  static const Color n0 = Color(0xFFFFFFFF);
  static const Color n25 = Color(0xFFF7F8F8);
  static const Color n50 = Color(0xFFF1F3F3);
  static const Color n100 = Color(0xFFE6E9EA);
  static const Color n200 = Color(0xFFCDD3D5);
  static const Color n300 = Color(0xFFA8B1B4);
  static const Color n400 = Color(0xFF7E898D);
  static const Color n500 = Color(0xFF5C686C);
  static const Color n600 = Color(0xFF3F4A4E);
  static const Color n700 = Color(0xFF2A3336);
  static const Color n800 = Color(0xFF1A2225);
  static const Color n900 = Color(0xFF0B1214);
  static const Color n950 = Color(0xFF070B0C);
}

/// Semantic status colors — light mode.
abstract final class NxSemanticColors {
  static const Color success100 = Color(0xFFE3F7EC);
  static const Color success600 = Color(0xFF1B7F4A);

  static const Color warning100 = Color(0xFFFFF4E0);
  static const Color warning600 = Color(0xFFB86E00);

  static const Color danger100 = Color(0xFFFDECEA);
  static const Color danger600 = Color(0xFFC62828);

  static const Color info100 = Color(0xFFEBF8F8);
  static const Color info600 = Color(0xFF0B6E6E);
}

/// Semantic status colors — dark mode.
abstract final class NxSemanticDarkColors {
  static const Color success100 = Color(0xFF143528);
  static const Color success600 = Color(0xFF3DDB8A);

  static const Color warning100 = Color(0xFF3A2A0A);
  static const Color warning600 = Color(0xFFFFC14D);

  static const Color danger100 = Color(0xFF3A1414);
  static const Color danger600 = Color(0xFFFF6B6B);

  static const Color info100 = Color(0xFF0E2C2C);
  static const Color info600 = Color(0xFF5ED0D0);
}

/// Dark-mode-only border tokens from semantic roles.
abstract final class NxDarkBorderColors {
  static const Color subtle = Color(0xFF243034);
  static const Color defaultBorder = Color(0xFF334044);
}

/// Resolved semantic color roles for theming.
@immutable
class NxColorRoles {
  const NxColorRoles({
    required this.bgCanvas,
    required this.bgSurface,
    required this.bgSurfaceRaised,
    required this.bgSunken,
    required this.bgNav,
    required this.bgBrand,
    required this.bgAccent,
    required this.bgDisabled,
    required this.bgOverlay,
    required this.borderSubtle,
    required this.borderDefault,
    required this.borderStrong,
    required this.borderFocus,
    required this.borderDanger,
    required this.textPrimary,
    required this.textSecondary,
    required this.textTertiary,
    required this.textDisabled,
    required this.textInverse,
    required this.textBrand,
    required this.textOnBrand,
    required this.textOnAccent,
    required this.textLink,
    required this.iconPrimary,
    required this.iconSecondary,
    required this.iconBrand,
    required this.navItemDefault,
    required this.navItemActive,
    required this.navIndicator,
    required this.success,
    required this.successSurface,
    required this.warning,
    required this.warningSurface,
    required this.danger,
    required this.dangerSurface,
    required this.info,
    required this.infoSurface,
  });

  final Color bgCanvas;
  final Color bgSurface;
  final Color bgSurfaceRaised;
  final Color bgSunken;
  final Color bgNav;
  final Color bgBrand;
  final Color bgAccent;
  final Color bgDisabled;
  final Color bgOverlay;
  final Color borderSubtle;
  final Color borderDefault;
  final Color borderStrong;
  final Color borderFocus;
  final Color borderDanger;
  final Color textPrimary;
  final Color textSecondary;
  final Color textTertiary;
  final Color textDisabled;
  final Color textInverse;
  final Color textBrand;
  final Color textOnBrand;
  final Color textOnAccent;
  final Color textLink;
  final Color iconPrimary;
  final Color iconSecondary;
  final Color iconBrand;
  final Color navItemDefault;
  final Color navItemActive;
  final Color navIndicator;
  final Color success;
  final Color successSurface;
  final Color warning;
  final Color warningSurface;
  final Color danger;
  final Color dangerSurface;
  final Color info;
  final Color infoSurface;

  static const NxColorRoles light = NxColorRoles(
    bgCanvas: NxNeutralColors.n25,
    bgSurface: NxNeutralColors.n0,
    bgSurfaceRaised: NxNeutralColors.n0,
    bgSunken: NxNeutralColors.n50,
    bgNav: NxNeutralColors.n0,
    bgBrand: NxBrandColors.primary600,
    bgAccent: NxBrandColors.accent400,
    bgDisabled: NxNeutralColors.n100,
    bgOverlay: Color(0x7A0B1214),
    borderSubtle: NxNeutralColors.n100,
    borderDefault: NxNeutralColors.n200,
    borderStrong: NxNeutralColors.n300,
    borderFocus: NxBrandColors.primary600,
    borderDanger: NxSemanticColors.danger600,
    textPrimary: NxNeutralColors.n700,
    textSecondary: NxNeutralColors.n500,
    textTertiary: NxNeutralColors.n400,
    textDisabled: NxNeutralColors.n300,
    textInverse: NxNeutralColors.n0,
    textBrand: NxBrandColors.primary700,
    textOnBrand: NxNeutralColors.n0,
    textOnAccent: NxBrandColors.ink,
    textLink: NxBrandColors.primary600,
    iconPrimary: NxNeutralColors.n700,
    iconSecondary: NxNeutralColors.n500,
    iconBrand: NxBrandColors.primary600,
    navItemDefault: NxNeutralColors.n500,
    navItemActive: NxBrandColors.primary600,
    navIndicator: NxBrandColors.primary600,
    success: NxSemanticColors.success600,
    successSurface: NxSemanticColors.success100,
    warning: NxSemanticColors.warning600,
    warningSurface: NxSemanticColors.warning100,
    danger: NxSemanticColors.danger600,
    dangerSurface: NxSemanticColors.danger100,
    info: NxSemanticColors.info600,
    infoSurface: NxSemanticColors.info100,
  );

  static const NxColorRoles dark = NxColorRoles(
    bgCanvas: NxNeutralColors.n950,
    bgSurface: NxNeutralColors.n900,
    bgSurfaceRaised: NxNeutralColors.n800,
    bgSunken: NxNeutralColors.n950,
    bgNav: NxNeutralColors.n900,
    bgBrand: NxBrandColors.primary600,
    bgAccent: NxBrandColors.accent400,
    bgDisabled: NxNeutralColors.n800,
    bgOverlay: Color(0x99000000),
    borderSubtle: NxDarkBorderColors.subtle,
    borderDefault: NxDarkBorderColors.defaultBorder,
    borderStrong: NxNeutralColors.n500,
    borderFocus: NxBrandColors.primary500,
    borderDanger: NxSemanticDarkColors.danger600,
    textPrimary: NxNeutralColors.n50,
    textSecondary: NxNeutralColors.n300,
    textTertiary: NxNeutralColors.n400,
    textDisabled: NxNeutralColors.n500,
    textInverse: NxNeutralColors.n900,
    textBrand: NxBrandColors.primary500,
    textOnBrand: NxNeutralColors.n0,
    textOnAccent: NxBrandColors.ink,
    textLink: NxBrandColors.primary500,
    iconPrimary: NxNeutralColors.n50,
    iconSecondary: NxNeutralColors.n300,
    iconBrand: NxBrandColors.primary500,
    navItemDefault: NxNeutralColors.n400,
    navItemActive: NxBrandColors.accent400,
    navIndicator: NxBrandColors.accent400,
    success: NxSemanticDarkColors.success600,
    successSurface: NxSemanticDarkColors.success100,
    warning: NxSemanticDarkColors.warning600,
    warningSurface: NxSemanticDarkColors.warning100,
    danger: NxSemanticDarkColors.danger600,
    dangerSurface: NxSemanticDarkColors.danger100,
    info: NxSemanticDarkColors.info600,
    infoSurface: NxSemanticDarkColors.info100,
  );

  NxColorRoles copyWith({
    Color? bgCanvas,
    Color? bgSurface,
    Color? bgSurfaceRaised,
    Color? bgSunken,
    Color? bgNav,
    Color? bgBrand,
    Color? bgAccent,
    Color? bgDisabled,
    Color? bgOverlay,
    Color? borderSubtle,
    Color? borderDefault,
    Color? borderStrong,
    Color? borderFocus,
    Color? borderDanger,
    Color? textPrimary,
    Color? textSecondary,
    Color? textTertiary,
    Color? textDisabled,
    Color? textInverse,
    Color? textBrand,
    Color? textOnBrand,
    Color? textOnAccent,
    Color? textLink,
    Color? iconPrimary,
    Color? iconSecondary,
    Color? iconBrand,
    Color? navItemDefault,
    Color? navItemActive,
    Color? navIndicator,
    Color? success,
    Color? successSurface,
    Color? warning,
    Color? warningSurface,
    Color? danger,
    Color? dangerSurface,
    Color? info,
    Color? infoSurface,
  }) {
    return NxColorRoles(
      bgCanvas: bgCanvas ?? this.bgCanvas,
      bgSurface: bgSurface ?? this.bgSurface,
      bgSurfaceRaised: bgSurfaceRaised ?? this.bgSurfaceRaised,
      bgSunken: bgSunken ?? this.bgSunken,
      bgNav: bgNav ?? this.bgNav,
      bgBrand: bgBrand ?? this.bgBrand,
      bgAccent: bgAccent ?? this.bgAccent,
      bgDisabled: bgDisabled ?? this.bgDisabled,
      bgOverlay: bgOverlay ?? this.bgOverlay,
      borderSubtle: borderSubtle ?? this.borderSubtle,
      borderDefault: borderDefault ?? this.borderDefault,
      borderStrong: borderStrong ?? this.borderStrong,
      borderFocus: borderFocus ?? this.borderFocus,
      borderDanger: borderDanger ?? this.borderDanger,
      textPrimary: textPrimary ?? this.textPrimary,
      textSecondary: textSecondary ?? this.textSecondary,
      textTertiary: textTertiary ?? this.textTertiary,
      textDisabled: textDisabled ?? this.textDisabled,
      textInverse: textInverse ?? this.textInverse,
      textBrand: textBrand ?? this.textBrand,
      textOnBrand: textOnBrand ?? this.textOnBrand,
      textOnAccent: textOnAccent ?? this.textOnAccent,
      textLink: textLink ?? this.textLink,
      iconPrimary: iconPrimary ?? this.iconPrimary,
      iconSecondary: iconSecondary ?? this.iconSecondary,
      iconBrand: iconBrand ?? this.iconBrand,
      navItemDefault: navItemDefault ?? this.navItemDefault,
      navItemActive: navItemActive ?? this.navItemActive,
      navIndicator: navIndicator ?? this.navIndicator,
      success: success ?? this.success,
      successSurface: successSurface ?? this.successSurface,
      warning: warning ?? this.warning,
      warningSurface: warningSurface ?? this.warningSurface,
      danger: danger ?? this.danger,
      dangerSurface: dangerSurface ?? this.dangerSurface,
      info: info ?? this.info,
      infoSurface: infoSurface ?? this.infoSurface,
    );
  }

  NxColorRoles lerp(NxColorRoles other, double t) {
    return NxColorRoles(
      bgCanvas: Color.lerp(bgCanvas, other.bgCanvas, t)!,
      bgSurface: Color.lerp(bgSurface, other.bgSurface, t)!,
      bgSurfaceRaised: Color.lerp(bgSurfaceRaised, other.bgSurfaceRaised, t)!,
      bgSunken: Color.lerp(bgSunken, other.bgSunken, t)!,
      bgNav: Color.lerp(bgNav, other.bgNav, t)!,
      bgBrand: Color.lerp(bgBrand, other.bgBrand, t)!,
      bgAccent: Color.lerp(bgAccent, other.bgAccent, t)!,
      bgDisabled: Color.lerp(bgDisabled, other.bgDisabled, t)!,
      bgOverlay: Color.lerp(bgOverlay, other.bgOverlay, t)!,
      borderSubtle: Color.lerp(borderSubtle, other.borderSubtle, t)!,
      borderDefault: Color.lerp(borderDefault, other.borderDefault, t)!,
      borderStrong: Color.lerp(borderStrong, other.borderStrong, t)!,
      borderFocus: Color.lerp(borderFocus, other.borderFocus, t)!,
      borderDanger: Color.lerp(borderDanger, other.borderDanger, t)!,
      textPrimary: Color.lerp(textPrimary, other.textPrimary, t)!,
      textSecondary: Color.lerp(textSecondary, other.textSecondary, t)!,
      textTertiary: Color.lerp(textTertiary, other.textTertiary, t)!,
      textDisabled: Color.lerp(textDisabled, other.textDisabled, t)!,
      textInverse: Color.lerp(textInverse, other.textInverse, t)!,
      textBrand: Color.lerp(textBrand, other.textBrand, t)!,
      textOnBrand: Color.lerp(textOnBrand, other.textOnBrand, t)!,
      textOnAccent: Color.lerp(textOnAccent, other.textOnAccent, t)!,
      textLink: Color.lerp(textLink, other.textLink, t)!,
      iconPrimary: Color.lerp(iconPrimary, other.iconPrimary, t)!,
      iconSecondary: Color.lerp(iconSecondary, other.iconSecondary, t)!,
      iconBrand: Color.lerp(iconBrand, other.iconBrand, t)!,
      navItemDefault: Color.lerp(navItemDefault, other.navItemDefault, t)!,
      navItemActive: Color.lerp(navItemActive, other.navItemActive, t)!,
      navIndicator: Color.lerp(navIndicator, other.navIndicator, t)!,
      success: Color.lerp(success, other.success, t)!,
      successSurface: Color.lerp(successSurface, other.successSurface, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      warningSurface: Color.lerp(warningSurface, other.warningSurface, t)!,
      danger: Color.lerp(danger, other.danger, t)!,
      dangerSurface: Color.lerp(dangerSurface, other.dangerSurface, t)!,
      info: Color.lerp(info, other.info, t)!,
      infoSurface: Color.lerp(infoSurface, other.infoSurface, t)!,
    );
  }
}
