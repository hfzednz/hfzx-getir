import 'package:nexora_core/nexora_core.dart';

/// Warehouse analytics event names.
abstract final class WarehouseAnalyticsEvents {
  static const version = '1';

  static const authOtpRequested = 'warehouse_auth_otp_requested';
  static const authOtpVerified = 'warehouse_auth_otp_verified';
  static const authLoginFailed = 'warehouse_auth_login_failed';
  static const shiftClockIn = 'warehouse_shift_clock_in';
  static const shiftClockOut = 'warehouse_shift_clock_out';

  static const pickTaskClaimed = 'warehouse_pick_task_claimed';
  static const pickLineScanned = 'warehouse_pick_line_scanned';
  static const pickShortPick = 'warehouse_pick_short_pick';
  static const pickCompleted = 'warehouse_pick_completed';

  static const packClaimed = 'warehouse_pack_claimed';
  static const packWeighed = 'warehouse_pack_weighed';
  static const packLabeled = 'warehouse_pack_labeled';
  static const packSealed = 'warehouse_pack_sealed';

  static const handoffVerified = 'warehouse_handoff_verified';
  static const handoffFailed = 'warehouse_handoff_failed';

  static const inventoryAdjusted = 'warehouse_inventory_adjusted';
  static const cycleCountSubmitted = 'warehouse_cycle_count_submitted';
}

/// Thin tracker over [AnalyticsGateway] with warehouse-safe props.
class WarehouseAnalyticsTracker {
  WarehouseAnalyticsTracker({
    required AnalyticsGateway gateway,
    this.storeIdProvider,
    this.userIdProvider,
    this.sessionIdProvider,
  }) : _gateway = gateway;

  final AnalyticsGateway _gateway;
  final String? Function()? storeIdProvider;
  final String? Function()? userIdProvider;
  final String? Function()? sessionIdProvider;

  Future<void> track({
    required String eventName,
    Map<String, Object?> props = const {},
    String eventVersion = WarehouseAnalyticsEvents.version,
  }) =>
      _gateway.track(
        eventName: eventName,
        eventVersion: eventVersion,
        props: {
          if (userIdProvider?.call() != null) 'user_id': userIdProvider!(),
          if (storeIdProvider?.call() != null) 'store_id': storeIdProvider!(),
          ...props,
        },
        cityId: storeIdProvider?.call(),
        sessionId: sessionIdProvider?.call(),
      );

  Future<void> trackPickClaimed({required String taskId}) => track(
        eventName: WarehouseAnalyticsEvents.pickTaskClaimed,
        props: {'task_id': taskId},
      );

  Future<void> trackPickLineScanned({
    required String taskId,
    required String sku,
    required String binId,
  }) =>
      track(
        eventName: WarehouseAnalyticsEvents.pickLineScanned,
        props: {
          'task_id': taskId,
          'sku': sku,
          'bin_id': binId,
        },
      );

  Future<void> trackHandoffVerified({
    required String handoffId,
    required String orderId,
    required String courierId,
  }) =>
      track(
        eventName: WarehouseAnalyticsEvents.handoffVerified,
        props: {
          'handoff_id': handoffId,
          'order_id': orderId,
          'courier_id': courierId,
        },
      );

  Future<void> flush() => _gateway.flush();
}
