/// 4pt grid spacing tokens from [nexora.tokens.json].
abstract final class NxSpacing {
  static const double s0 = 0;
  static const double s0_5 = 2;
  static const double s1 = 4;
  static const double s2 = 8;
  static const double s3 = 12;
  static const double s4 = 16;
  static const double s5 = 20;
  static const double s6 = 24;
  static const double s8 = 32;
  static const double s10 = 40;
  static const double s12 = 48;
  static const double s16 = 64;
  static const double s20 = 80;
  static const double s24 = 96;
}

/// Border width tokens.
abstract final class NxBorderWidth {
  static const double hairline = 1;
  static const double thick = 2;
  static const double heavy = 3;
}

/// Icon size tokens.
abstract final class NxIconSize {
  static const double xs = 12;
  static const double sm = 16;
  static const double md = 20;
  static const double lg = 24;
  static const double xl = 32;
  static const double xxl = 48;
}

/// Z-index layering tokens.
abstract final class NxZIndex {
  static const int base = 0;
  static const int sticky = 10;
  static const int nav = 20;
  static const int dropdown = 30;
  static const int overlay = 40;
  static const int modal = 50;
  static const int toast = 60;
  static const int commandPalette = 70;
}
