import 'package:flutter/animation.dart';

/// Motion duration tokens (milliseconds).
abstract final class NxDuration {
  static const Duration instant = Duration.zero;
  static const Duration fast = Duration(milliseconds: 120);
  static const Duration micro = Duration(milliseconds: 180);
  static const Duration short = Duration(milliseconds: 260);
  static const Duration medium = Duration(milliseconds: 320);
  static const Duration long = Duration(milliseconds: 440);
  static const Duration xlong = Duration(milliseconds: 600);

  /// Signature motion — cart commit morph.
  static const Duration cartCommit = Duration(milliseconds: 320);

  /// Signature motion — order pulse ring.
  static const Duration orderPulse = Duration(milliseconds: 400);

  /// Signature motion — ETA breathe cycle.
  static const Duration etaBreathe = Duration(milliseconds: 1600);
}

/// Motion easing curves from token cubic-bezier values.
abstract final class NxCurves {
  static const Curve standard = Cubic(0.2, 0.0, 0.0, 1.0);
  static const Curve emphasized = Cubic(0.2, 0.0, 0.0, 1.0);
  static const Curve decelerate = Cubic(0.0, 0.0, 0.2, 1.0);
  static const Curve accelerate = Cubic(0.3, 0.0, 1.0, 1.0);
}
