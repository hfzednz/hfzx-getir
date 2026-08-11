import 'package:flutter/material.dart';

import '../tokens/colors.dart';
import '../tokens/radius.dart';

/// Shimmer skeleton placeholder — 12° gradient, 1.2s loop, neutrals only.
class NxSkeleton extends StatefulWidget {
  const NxSkeleton({
    super.key,
    this.width,
    this.height = 16,
    this.borderRadius,
  });

  final double? width;
  final double height;
  final BorderRadius? borderRadius;

  @override
  State<NxSkeleton> createState() => _NxSkeletonState();
}

class _NxSkeletonState extends State<NxSkeleton> with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final disableAnimations = MediaQuery.disableAnimationsOf(context);
    final radius = widget.borderRadius ?? BorderRadius.circular(NxRadius.sm);

    if (disableAnimations) {
      return Container(
        width: widget.width,
        height: widget.height,
        decoration: BoxDecoration(
          color: NxNeutralColors.n100,
          borderRadius: radius,
        ),
      );
    }

    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Container(
          width: widget.width,
          height: widget.height,
          decoration: BoxDecoration(
            borderRadius: radius,
            gradient: LinearGradient(
              begin: Alignment(-1.0 + _controller.value * 2, -0.2),
              end: Alignment(1.0 + _controller.value * 2, 0.2),
              colors: const [
                NxNeutralColors.n100,
                NxNeutralColors.n50,
                NxNeutralColors.n100,
              ],
              stops: const [0.25, 0.5, 0.75],
              transform: const GradientRotation(0.21),
            ),
          ),
        );
      },
    );
  }
}
