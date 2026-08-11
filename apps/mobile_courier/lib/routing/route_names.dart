/// Route path constants for GoRouter (ARCHITECTURE.md).
abstract final class RouteNames {
  static const splash = '/';
  static const auth = '/auth';
  static const authPhone = '/auth/phone';
  static const authOtp = '/auth/otp';
  static const authKyc = '/auth/kyc';
  static const shell = '/shell';
  static const home = '/home';
  static const offers = '/offers';
  static const deliveries = '/deliveries';
  static const earnings = '/earnings';
  static const account = '/account';
  static const deliveryDetail = '/deliveries/:id';
  static const deliveryNavigate = '/deliveries/:id/navigate';
  static const deliveryPickup = '/deliveries/:id/pickup';
  static const deliveryPod = '/deliveries/:id/pod';
  static const deliveryFailed = '/deliveries/:id/failed';
  static const shifts = '/shifts';
  static const performance = '/performance';
  static const profile = '/profile';
  static const documents = '/profile/documents';
  static const support = '/support';
  static const settings = '/settings';
  static const notifications = '/notifications';
  static const status = '/status';

  static String deliveryDetailPath(String id) => '/deliveries/$id';
  static String deliveryNavigatePath(String id) => '/deliveries/$id/navigate';
  static String deliveryPickupPath(String id) => '/deliveries/$id/pickup';
  static String deliveryPodPath(String id) => '/deliveries/$id/pod';
  static String deliveryFailedPath(String id) => '/deliveries/$id/failed';
}

/// Paths requiring authenticated courier session.
const authRequiredPrefixes = [
  RouteNames.home,
  RouteNames.offers,
  RouteNames.deliveries,
  RouteNames.earnings,
  RouteNames.account,
  RouteNames.shifts,
  RouteNames.performance,
  RouteNames.profile,
  RouteNames.support,
  RouteNames.settings,
  RouteNames.notifications,
  '/shell',
];
