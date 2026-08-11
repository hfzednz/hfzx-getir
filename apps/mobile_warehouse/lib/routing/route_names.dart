/// Route path constants for GoRouter (ARCHITECTURE.md).
abstract final class RouteNames {
  static const splash = '/';
  static const auth = '/auth';
  static const authPhone = '/auth/phone';
  static const authOtp = '/auth/otp';
  static const home = '/home';
  static const picking = '/picking';
  static const packing = '/packing';
  static const dispatch = '/dispatch';
  static const more = '/more';

  static const inventory = '/inventory';
  static const cycleCount = '/inventory/cycle-count';
  static const inboundReceive = '/inventory/inbound';
  static const transfers = '/transfers';
  static const transferCreate = '/transfers/create';
  static const expiry = '/expiry';
  static const purchasing = '/purchasing';
  static const returns = '/returns';
  static const quality = '/quality';
  static const map = '/map';
  static const ai = '/ai';
  static const shifts = '/shifts';
  static const tasks = '/tasks';
  static const reports = '/reports';
  static const notifications = '/notifications';
  static const settings = '/settings';
  static const support = '/support';

  static String pickingTaskPath(String id) => '/picking/$id';
  static String pickingScanPath(String id) => '/picking/$id/scan';
  static String packingTaskPath(String id) => '/packing/$id';
  static String dispatchHandoffPath(String id) => '/dispatch/$id';
  static String dispatchScanPath(String id) => '/dispatch/$id/scan';
  static String inventoryAdjustPath(String sku) => '/inventory/adjust/$sku';
}

/// Paths requiring authenticated warehouse session.
const authRequiredPrefixes = [
  RouteNames.home,
  RouteNames.picking,
  RouteNames.packing,
  RouteNames.dispatch,
  RouteNames.more,
  RouteNames.inventory,
  RouteNames.transfers,
  RouteNames.expiry,
  RouteNames.purchasing,
  RouteNames.returns,
  RouteNames.quality,
  RouteNames.map,
  RouteNames.ai,
  RouteNames.shifts,
  RouteNames.tasks,
  RouteNames.reports,
  RouteNames.notifications,
  RouteNames.settings,
  RouteNames.support,
];
