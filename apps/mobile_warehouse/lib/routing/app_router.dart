import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/ai/presentation/screens/ai_hub_screen.dart';
import '../features/auth/presentation/screens/auth_otp_screen.dart';
import '../features/auth/presentation/screens/auth_phone_screen.dart';
import '../features/dispatch/presentation/screens/dispatch_queue_screen.dart';
import '../features/dispatch/presentation/screens/handoff_detail_screen.dart';
import '../features/dispatch/presentation/screens/handoff_scan_screen.dart';
import '../features/expiry/presentation/screens/expiry_screen.dart';
import '../features/home/presentation/screens/home_screen.dart';
import '../features/inventory/presentation/screens/adjust_stock_screen.dart';
import '../features/inventory/presentation/screens/cycle_count_screen.dart';
import '../features/inventory/presentation/screens/inbound_receive_screen.dart';
import '../features/inventory/presentation/screens/inventory_screen.dart';
import '../features/map/presentation/screens/warehouse_map_screen.dart';
import '../features/more/presentation/screens/more_screen.dart';
import '../features/notifications/presentation/screens/notifications_screen.dart';
import '../features/packing/presentation/screens/pack_task_screen.dart';
import '../features/packing/presentation/screens/packing_queue_screen.dart';
import '../features/picking/presentation/screens/pick_scan_screen.dart';
import '../features/picking/presentation/screens/pick_task_screen.dart';
import '../features/picking/presentation/screens/picking_queue_screen.dart';
import '../features/purchasing/presentation/screens/purchasing_screen.dart';
import '../features/quality/presentation/screens/quality_screen.dart';
import '../features/reports/presentation/screens/reports_screen.dart';
import '../features/returns/presentation/screens/returns_screen.dart';
import '../features/settings/presentation/screens/settings_screen.dart';
import '../features/shell/presentation/screens/shell_scaffold.dart';
import '../features/shifts/presentation/screens/shifts_screen.dart';
import '../features/splash/presentation/screens/splash_screen.dart';
import '../features/support/presentation/screens/support_screen.dart';
import '../features/tasks/presentation/screens/tasks_screen.dart';
import '../features/transfers/presentation/screens/create_transfer_screen.dart';
import '../features/transfers/presentation/screens/transfers_screen.dart';
import 'auth_guard.dart';
import 'route_names.dart';

final _rootNavigatorKey = GlobalKey<NavigatorState>();
final _shellHomeKey = GlobalKey<NavigatorState>(debugLabel: 'home');
final _shellPickKey = GlobalKey<NavigatorState>(debugLabel: 'picking');
final _shellPackKey = GlobalKey<NavigatorState>(debugLabel: 'packing');
final _shellDispatchKey = GlobalKey<NavigatorState>(debugLabel: 'dispatch');
final _shellMoreKey = GlobalKey<NavigatorState>(debugLabel: 'more');

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
        ],
      ),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) {
          return ShellScaffold(navigationShell: navigationShell);
        },
        branches: [
          StatefulShellBranch(
            navigatorKey: _shellHomeKey,
            routes: [
              GoRoute(
                path: RouteNames.home,
                builder: (_, __) => const HomeScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellPickKey,
            routes: [
              GoRoute(
                path: RouteNames.picking,
                builder: (_, __) => const PickingQueueScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellPackKey,
            routes: [
              GoRoute(
                path: RouteNames.packing,
                builder: (_, __) => const PackingQueueScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellDispatchKey,
            routes: [
              GoRoute(
                path: RouteNames.dispatch,
                builder: (_, __) => const DispatchQueueScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellMoreKey,
            routes: [
              GoRoute(
                path: RouteNames.more,
                builder: (_, __) => const MoreScreen(),
              ),
            ],
          ),
        ],
      ),
      GoRoute(
        path: '/picking/:taskId',
        parentNavigatorKey: _rootNavigatorKey,
        builder: (_, state) => PickTaskScreen(
          taskId: state.pathParameters['taskId']!,
        ),
        routes: [
          GoRoute(
            path: 'scan',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => PickScanScreen(
              taskId: state.pathParameters['taskId']!,
            ),
          ),
        ],
      ),
      GoRoute(
        path: '/packing/:taskId',
        parentNavigatorKey: _rootNavigatorKey,
        builder: (_, state) => PackTaskScreen(
          taskId: state.pathParameters['taskId']!,
        ),
      ),
      GoRoute(
        path: '/dispatch/:handoffId',
        parentNavigatorKey: _rootNavigatorKey,
        builder: (_, state) => HandoffDetailScreen(
          handoffId: state.pathParameters['handoffId']!,
        ),
        routes: [
          GoRoute(
            path: 'scan',
            parentNavigatorKey: _rootNavigatorKey,
            builder: (_, state) => HandoffScanScreen(
              handoffId: state.pathParameters['handoffId']!,
            ),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.inventory,
        builder: (_, __) => const InventoryScreen(),
        routes: [
          GoRoute(
            path: 'cycle-count',
            builder: (_, __) => const CycleCountScreen(),
          ),
          GoRoute(
            path: 'inbound',
            builder: (_, __) => const InboundReceiveScreen(),
          ),
          GoRoute(
            path: 'adjust/:sku',
            builder: (_, state) => AdjustStockScreen(
              sku: state.pathParameters['sku']!,
            ),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.transfers,
        builder: (_, __) => const TransfersScreen(),
        routes: [
          GoRoute(
            path: 'create',
            builder: (_, __) => const CreateTransferScreen(),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.expiry,
        builder: (_, __) => const ExpiryScreen(),
      ),
      GoRoute(
        path: RouteNames.purchasing,
        builder: (_, __) => const PurchasingScreen(),
      ),
      GoRoute(
        path: RouteNames.returns,
        builder: (_, __) => const ReturnsScreen(),
      ),
      GoRoute(
        path: RouteNames.quality,
        builder: (_, __) => const QualityScreen(),
      ),
      GoRoute(
        path: RouteNames.map,
        builder: (_, __) => const WarehouseMapScreen(),
      ),
      GoRoute(
        path: RouteNames.ai,
        builder: (_, __) => const AiHubScreen(),
      ),
      GoRoute(
        path: RouteNames.shifts,
        builder: (_, __) => const ShiftsScreen(),
      ),
      GoRoute(
        path: RouteNames.tasks,
        builder: (_, __) => const TasksScreen(),
      ),
      GoRoute(
        path: RouteNames.reports,
        builder: (_, __) => const ReportsScreen(),
      ),
      GoRoute(
        path: RouteNames.notifications,
        builder: (_, __) => const NotificationsScreen(),
      ),
      GoRoute(
        path: RouteNames.settings,
        builder: (_, __) => const SettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.support,
        builder: (_, __) => const SupportScreen(),
      ),
    ],
  );
});
