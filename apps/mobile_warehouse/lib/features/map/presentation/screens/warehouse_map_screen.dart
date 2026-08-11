import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/warehouse_layout.dart';
import '../providers/map_providers.dart';

class WarehouseMapScreen extends ConsumerWidget {
  const WarehouseMapScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(warehouseMapProvider);
    final colors = context.nxColors;
    return Scaffold(
      appBar: const NxTopBar(title: 'Warehouse map'),
      body: AsyncValueWidget<WarehouseLayout>(
        value: async,
        data: (layout) => Column(
          children: [
            Expanded(
              flex: 3,
              child: Padding(
                padding: const EdgeInsets.all(NxSpacing.s3),
                child: CustomPaint(
                  painter: _LayoutPainter(layout.zones, colors.borderDefault, colors.bgAccent),
                  child: const SizedBox.expand(),
                ),
              ),
            ),
            Expanded(
              flex: 2,
              child: ListView.separated(
                padding: const EdgeInsets.all(NxSpacing.s3),
                itemCount: layout.zones.length,
                separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
                itemBuilder: (context, i) {
                  final z = layout.zones[i];
                  return NxCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(z.name, style: NxTypography.titleSm),
                        Text('Aisles: ${z.aisles.join(', ')}', style: NxTypography.captionMd),
                      ],
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _LayoutPainter extends CustomPainter {
  _LayoutPainter(this.zones, this.border, this.fill);
  final List<WarehouseZone> zones;
  final Color border;
  final Color fill;

  @override
  void paint(Canvas canvas, Size size) {
    if (zones.isEmpty) return;
    double maxX = 1, maxY = 1;
    for (final z in zones) {
      maxX = maxX < z.x + z.width ? z.x + z.width : maxX;
      maxY = maxY < z.y + z.height ? z.y + z.height : maxY;
    }
    final sx = size.width / maxX;
    final sy = size.height / maxY;
    final paint = Paint()..color = fill.withValues(alpha: 0.35);
    final stroke = Paint()
      ..color = border
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5;
    for (final z in zones) {
      final rect = Rect.fromLTWH(z.x * sx, z.y * sy, z.width * sx, z.height * sy);
      canvas.drawRRect(RRect.fromRectAndRadius(rect, const Radius.circular(4)), paint);
      canvas.drawRRect(RRect.fromRectAndRadius(rect, const Radius.circular(4)), stroke);
    }
  }

  @override
  bool shouldRepaint(covariant _LayoutPainter oldDelegate) =>
      oldDelegate.zones != zones;
}
