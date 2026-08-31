import '../../orders/domain/entities/orders_entity.dart';
import '../domain/entities/home_entity.dart';

List<HomeWidgetConfig> historyWidgetsFromOrders(List<Order> orders) {
  final recent = <HomeProduct>[];
  final seen = <String>{};
  final counts = <String, int>{};
  final tiles = <String, HomeProduct>{};
  for (final order in orders) {
    for (final line in order.items) {
      final id = line.productId.isNotEmpty ? line.productId : line.sku ?? '';
      if (id.isEmpty) continue;
      final tile = HomeProduct(
        id: id,
        title: line.name.isNotEmpty ? line.name : id,
        priceMinor: line.unitPriceMinor,
        deepLink: '/p/$id',
      );
      tiles[id] = tile;
      counts[id] = (counts[id] ?? 0) + line.quantity;
      if (seen.add(id) && recent.length < 12) {
        recent.add(tile);
      }
    }
  }
  final widgets = <HomeWidgetConfig>[];
  widgets.add(
    HomeWidgetConfig(
      id: 'recently-ordered',
      type: HomeWidgetType.recentlyViewed,
      title: 'Recently ordered',
      items: recent,
    ),
  );
  final frequentIds = counts.entries.toList()
    ..sort((a, b) => b.value.compareTo(a.value));
  final frequent = [
    for (final entry in frequentIds)
      if (tiles[entry.key] != null) tiles[entry.key]!,
  ].take(8).toList();
  widgets.add(
    HomeWidgetConfig(
      id: 'frequently-purchased',
      type: HomeWidgetType.recommendation,
      title: 'Frequently purchased',
      items: frequent,
    ),
  );
  return widgets;
}

List<HomeWidgetConfig> mergeHomeHistory(
  List<HomeWidgetConfig> widgets,
  List<Order> orders,
) {
  final without = widgets
      .where((w) => w.id != 'recently-ordered' && w.id != 'frequently-purchased')
      .toList();
  return [...without, ...historyWidgetsFromOrders(orders)];
}
