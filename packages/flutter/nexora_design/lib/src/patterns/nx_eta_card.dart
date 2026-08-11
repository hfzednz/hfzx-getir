import 'package:flutter/material.dart';

import '../components/nx_card.dart';
import '../theme/nx_theme.dart';
import '../tokens/motion.dart';
import '../tokens/spacing.dart';
import '../tokens/typography.dart';

/// ETA card with optional signature breathe animation.
class NxEtaCard extends StatefulWidget {
  const NxEtaCard({
    super.key,
    required this.etaRange,
    this.confidenceCopy,
    this.storeName,
    this.live = false,
    this.semanticLabel,
  });

  final String etaRange;
  final String? confidenceCopy;
  final String? storeName;
  final bool live;
  final String? semanticLabel;

  @override
  State<NxEtaCard> createState() => _NxEtaCardState();
}

class _NxEtaCardState extends State<NxEtaCard> with SingleTickerProviderStateMixin {
  AnimationController? _breatheController;
  Animation<double>? _scale;
  Animation<double>? _opacity;

  @override
  void initState() {
    super.initState();
    _setupAnimation();
  }

  @override
  void didUpdateWidget(covariant NxEtaCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.live != widget.live) {
      _disposeAnimation();
      _setupAnimation();
    }
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final disableAnimations = MediaQuery.disableAnimationsOf(context);
    if (widget.live && disableAnimations) {
      _disposeAnimation();
    } else if (widget.live && _breatheController == null && !disableAnimations) {
      _setupAnimation();
    }
  }

  void _setupAnimation() {
    if (!widget.live || _breatheController != null) return;
    _breatheController = AnimationController(
      vsync: this,
      duration: NxDuration.etaBreathe,
    )..repeat(reverse: true);

    _scale = Tween<double>(begin: 1.0, end: 1.02).animate(
      CurvedAnimation(parent: _breatheController!, curve: NxCurves.standard),
    );
    _opacity = Tween<double>(begin: 1.0, end: 0.92).animate(
      CurvedAnimation(parent: _breatheController!, curve: NxCurves.standard),
    );
  }

  void _disposeAnimation() {
    _breatheController?.dispose();
    _breatheController = null;
    _scale = null;
    _opacity = null;
  }

  @override
  void dispose() {
    _disposeAnimation();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    final disableAnimations = MediaQuery.disableAnimationsOf(context);
    final shouldAnimate = widget.live && !disableAnimations && _breatheController != null;

    Widget content = NxCard(
      variant: NxCardVariant.outlined,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (widget.storeName != null)
            Text(
              widget.storeName!,
              style: NxTypography.captionMd.copyWith(color: colors.textSecondary),
            ),
          Text(
            widget.etaRange,
            style: NxTypography.etaMd.copyWith(color: colors.textPrimary),
          ),
          if (widget.confidenceCopy != null) ...[
            const SizedBox(height: NxSpacing.s1),
            Text(
              widget.confidenceCopy!,
              style: NxTypography.bodySm.copyWith(color: colors.textTertiary),
            ),
          ],
        ],
      ),
    );

    content = Semantics(
      label: widget.semanticLabel ?? 'Estimated arrival ${widget.etaRange}',
      child: content,
    );

    if (!shouldAnimate) return content;

    return AnimatedBuilder(
      animation: _breatheController!,
      builder: (context, child) {
        return Opacity(
          opacity: _opacity!.value,
          child: Transform.scale(scale: _scale!.value, child: child),
        );
      },
      child: content,
    );
  }
}
