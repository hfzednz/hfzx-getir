import 'package:equatable/equatable.dart';

class WarehouseKpis extends Equatable {
  const WarehouseKpis({
    this.pickSpeedPerHour = 0,
    this.pickAccuracy = 0,
    this.packSpeedPerHour = 0,
    this.wasteUnits = 0,
    this.onTimeDispatchRate = 0,
  });

  final double pickSpeedPerHour;
  final double pickAccuracy;
  final double packSpeedPerHour;
  final int wasteUnits;
  final double onTimeDispatchRate;

  factory WarehouseKpis.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const WarehouseKpis();
    return WarehouseKpis(
      pickSpeedPerHour: (json['pick_speed_per_hour'] as num?)?.toDouble() ?? 0,
      pickAccuracy: (json['pick_accuracy'] as num?)?.toDouble() ?? 0,
      packSpeedPerHour: (json['pack_speed_per_hour'] as num?)?.toDouble() ?? 0,
      wasteUnits: (json['waste_units'] as num?)?.toInt() ?? 0,
      onTimeDispatchRate:
          (json['on_time_dispatch_rate'] as num?)?.toDouble() ?? 0,
    );
  }

  @override
  List<Object?> get props => [
        pickSpeedPerHour,
        pickAccuracy,
        packSpeedPerHour,
        wasteUnits,
        onTimeDispatchRate,
      ];
}

class WarehouseDashboard extends Equatable {
  const WarehouseDashboard({
    this.ordersWaiting = 0,
    this.pickQueue = 0,
    this.packQueue = 0,
    this.dispatchQueue = 0,
    this.courierArrivals = 0,
    this.lowStockAlerts = 0,
    this.oosAlerts = 0,
    this.expiryAlerts = 0,
    this.kpis = const WarehouseKpis(),
    this.aiTip,
  });

  final int ordersWaiting;
  final int pickQueue;
  final int packQueue;
  final int dispatchQueue;
  final int courierArrivals;
  final int lowStockAlerts;
  final int oosAlerts;
  final int expiryAlerts;
  final WarehouseKpis kpis;
  final String? aiTip;

  factory WarehouseDashboard.fromJson(Map<String, dynamic> json) {
    return WarehouseDashboard(
      ordersWaiting: (json['orders_waiting'] as num?)?.toInt() ?? 0,
      pickQueue: (json['pick_queue'] as num?)?.toInt() ?? 0,
      packQueue: (json['pack_queue'] as num?)?.toInt() ?? 0,
      dispatchQueue: (json['dispatch_queue'] as num?)?.toInt() ?? 0,
      courierArrivals: (json['courier_arrivals'] as num?)?.toInt() ?? 0,
      lowStockAlerts: (json['low_stock_alerts'] as num?)?.toInt() ?? 0,
      oosAlerts: (json['oos_alerts'] as num?)?.toInt() ?? 0,
      expiryAlerts: (json['expiry_alerts'] as num?)?.toInt() ?? 0,
      kpis: WarehouseKpis.fromJson(json['kpis'] as Map<String, dynamic>?),
      aiTip: json['ai_tip']?.toString() ?? json['tip']?.toString(),
    );
  }

  @override
  List<Object?> get props => [
        ordersWaiting,
        pickQueue,
        packQueue,
        dispatchQueue,
        courierArrivals,
        lowStockAlerts,
        oosAlerts,
        expiryAlerts,
        kpis,
        aiTip,
      ];
}
