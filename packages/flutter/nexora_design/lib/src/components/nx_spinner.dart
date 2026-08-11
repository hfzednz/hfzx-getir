import 'package:flutter/material.dart';

import '../theme/nx_theme.dart';

enum NxSpinnerSize { sm, md, lg }

/// Brand arc spinner — teal with optional accent endpoint.
class NxSpinner extends StatefulWidget {
  const NxSpinner({
    super.key,
    this.size = NxSpinnerSize.md,
    this.color,
    this.dimension,
  });

  final NxSpinnerSize size;
  final Color? color;
  final double? dimension;

  double get resolvedDimension => dimension ?? switch (size) {
        NxSpinnerSize.sm => 16,
        NxSpinnerSize.md => 24,
        NxSpinnerSize.lg => 32,
      };

  @override
  State<NxSpinner> createState() => _NxSpinnerState();
}

class _NxSpinnerState extends State<NxSpinner> with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final dimension = widget.resolvedDimension;
    final color = widget.color ?? colors.iconBrand;

    return Semantics(
      label: 'Loading',
      child: SizedBox(
        width: dimension,
        height: dimension,
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, child) {
            return CustomPaint(
              painter: _ArcSpinnerPainter(
                progress: _controller.value,
                color: color,
                accent: colors.bgAccent,
                strokeWidth: dimension <= 16 ? 2 : 2.5,
              ),
            );
          },
        ),
      ),
    );
  }
}

class _ArcSpinnerPainter extends CustomPainter {
  _ArcSpinnerPainter({
    required this.progress,
    required this.color,
    required this.accent,
    required this.strokeWidth,
  });

  final double progress;
  final Color color;
  final Color accent;
  final double strokeWidth;

  @override
  void paint(Canvas canvas, Size size) {
    final rect = Offset.zero & size;
    final trackPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..color = color.withValues(alpha: 0.2);

    canvas.drawArc(rect.deflate(strokeWidth), 0, 6.28, false, trackPaint);

    final arcPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round
      ..color = color;

    canvas.drawArc(
      rect.deflate(strokeWidth),
      progress * 6.28,
      2.4,
      false,
      arcPaint,
    );

    final accentPaint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round
      ..color = accent;

    canvas.drawArc(
      rect.deflate(strokeWidth),
      progress * 6.28 + 2.0,
      0.4,
      false,
      accentPaint,
    );
  }

  @override
  bool shouldRepaint(covariant _ArcSpinnerPainter oldDelegate) =>
      oldDelegate.progress != progress;
}
