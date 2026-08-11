import 'package:equatable/equatable.dart';

class WarehouseZone extends Equatable {
  const WarehouseZone({
    required this.id,
    required this.name,
    required this.aisles,
    this.x = 0,
    this.y = 0,
    this.width = 1,
    this.height = 1,
  });
  final String id;
  final String name;
  final List<String> aisles;
  final double x;
  final double y;
  final double width;
  final double height;
  factory WarehouseZone.fromJson(Map<String, dynamic> json) {
    final aisles = (json['aisles'] as List? ?? const []).map((e) => e.toString()).toList();
    return WarehouseZone(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      aisles: aisles,
      x: (json['x'] as num?)?.toDouble() ?? 0,
      y: (json['y'] as num?)?.toDouble() ?? 0,
      width: (json['width'] as num?)?.toDouble() ?? 1,
      height: (json['height'] as num?)?.toDouble() ?? 1,
    );
  }
  @override
  List<Object?> get props => [id, name, aisles, x, y, width, height];
}

class WarehouseLayout extends Equatable {
  const WarehouseLayout({required this.storeId, required this.zones});
  final String storeId;
  final List<WarehouseZone> zones;
  factory WarehouseLayout.fromJson(Map<String, dynamic> json) {
    final zones = (json['zones'] as List? ?? const [])
        .map((e) => WarehouseZone.fromJson(Map<String, dynamic>.from(e as Map)))
        .toList();
    return WarehouseLayout(
      storeId: json['store_id']?.toString() ?? '',
      zones: zones,
    );
  }
  @override
  List<Object?> get props => [storeId, zones];
}
