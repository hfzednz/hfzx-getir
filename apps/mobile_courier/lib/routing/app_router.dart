import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/auth/presentation/screens/auth_otp_screen.dart';
import '../features/auth/presentation/screens/auth_phone_screen.dart';
import '../features/auth/presentation/screens/kyc_screen.dart';
import '../features/deliveries/presentation/screens/deliveries_screen.dart';
import '../features/deliveries/presentation/screens/delivery_detail_screen.dart';
import '../features/deliveries/presentation/screens/delivery_pickup_screen.dart';
import '../features/deliveries/presentation/screens/delivery_pod_screen.dart';
import '../features/deliveries/presentation/screens/failed_delivery_screen.dart';
import '../features/earnings/presentation/screens/earnings_screen.dart';
import '../features/home/presentation/screens/home_screen.dart';
import '../features/navigation/presentation/screens/delivery_navigate_screen.dart';
import '../features/notifications/presentation/screens/notifications_screen.dart';
import '../features/offers/presentation/screens/offers_screen.dart';
import '../features/performance/presentation/screens/performance_screen.dart';
import '../features/profile/presentation/screens/account_screen.dart';
import '../features/profile/presentation/screens/documents_screen.dart';
import '../features/profile/presentation/screens/profile_screen.dart';
import '../features/settings/presentation/screens/settings_screen.dart';
import '../features/shell/presentation/screens/shell_scaffold.dart';
import '../features/shifts/presentation/screens/shifts_screen.dart';
import '../features/splash/presentation/screens/splash_screen.dart';
import '../features/status/presentation/screens/status_screen.dart';
import '../features/support/presentation/screens/support_screen.dart';
import 'auth_guard.dart';
import 'route_names.dart';

final _rootNavigatorKey = GlobalKey<NavigatorState>();
final _shellNavigatorHomeKey = GlobalKey<NavigatorState>(debugLabel: 'home');
final _shellNavigatorOffersKey = GlobalKey<NavigatorState>(debugLabel: 'offers');
final _shellNavigatorDeliveriesKey =
    GlobalKey<NavigatorState>(debugLabel: 'deliveries');
final _shellNavigatorEarningsKey =
    GlobalKey<NavigatorState>(debugLabel: 'earnings');
final _shellNavigatorAccountKey =
    GlobalKey<NavigatorState>(debugLabel: 'account');

final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    navigatorKey: _rootNavigatorKey,
    initialLocation: RouteNames.splash,
    redirect: (context, state) => authGuard(context, state, ref),
    routes: [
      GoRoute(
        path: RouteNames.splash,
        builder: (_, __) => const SplashScreen(),
      ),
      GoRoute(
        path: RouteNames.auth,
        redirect: (_, __) => RouteNames.authPhone,
        routes: [
          GoRoute(
            path: 'phone',
            builder: (_, __) => const AuthPhoneScreen(),
          ),
          GoRoute(
            path: 'otp',
            builder: (_, state) => AuthOtpScreen(
              phone: state.uri.queryParameters['phone'],
            ),
          ),
          GoRoute(
            path: 'kyc',
            builder: (_, __) => const KycScreen(),
          ),
        ],
      ),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) {
          return ShellScaffold(navigationShell: navigationShell);
        },
        branches: [
          StatefulShellBranch(
            navigatorKey: _shellNavigatorHomeKey,
            routes: [
              GoRoute(
                path: RouteNames.home,
                builder: (_, __) => const HomeScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorOffersKey,
            routes: [
              GoRoute(
                path: RouteNames.offers,
                builder: (_, __) => const OffersScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorDeliveriesKey,
            routes: [
              GoRoute(
                path: RouteNames.deliveries,
                builder: (_, __) => const DeliveriesScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorEarningsKey,
            routes: [
              GoRoute(
                path: RouteNames.earnings,
                builder: (_, __) => const EarningsScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorAccountKey,
            routes: [
              GoRoute(
                path: RouteNames.account,
                builder: (_, __) => const AccountScreen(),
              ),
            ],
          ),
        ],
      ),
      GoRoute(
        path: '/deliveries/:id',
        parentNavigatorKey: _rootNavigatorKey,
        builder: (_, state) => DeliveryDetailScreen(
          deliveryId: state.pathParameters['id']!,
        ),
        routes: [
          GoRoute(
            path: 'navigate',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => DeliveryNavigateScreen(
              deliveryId: state.pathParameters['id']!,
            ),
          ),
          GoRoute(
            path: 'pickup',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => DeliveryPickupScreen(
              deliveryId: state.pathParameters['id']!,
            ),
          ),
          GoRoute(
            path: 'pod',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => DeliveryPodScreen(
              deliveryId: state.pathParameters['id']!,
            ),
          ),
          GoRoute(
            path: 'failed',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => FailedDeliveryScreen(
              deliveryId: state.pathParameters['id']!,
            ),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.shifts,
        builder: (_, __) => const ShiftsScreen(),
      ),
      GoRoute(
        path: RouteNames.performance,
        builder: (_, __) => const PerformanceScreen(),
      ),
      GoRoute(
        path: RouteNames.profile,
        builder: (_, __) => const ProfileScreen(),
        routes: [
          GoRoute(
            path: 'documents',
            builder: (_, __) => const DocumentsScreen(),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.support,
        builder: (_, __) => const SupportScreen(),
      ),
      GoRoute(
        path: RouteNames.settings,
        builder: (_, __) => const SettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.notifications,
        builder: (_, __) => const NotificationsScreen(),
      ),
      GoRoute(
        path: RouteNames.status,
        builder: (_, __) => const StatusScreen(),
      ),
    ],
  );
});
