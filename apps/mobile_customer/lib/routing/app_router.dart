import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/addresses/presentation/screens/address_edit_screen.dart';
import '../features/addresses/presentation/screens/addresses_screen.dart';
import '../features/auth/presentation/screens/auth_email_screen.dart';
import '../features/auth/presentation/screens/auth_otp_screen.dart';
import '../features/auth/presentation/screens/auth_phone_screen.dart';
import '../features/auth/presentation/screens/auth_welcome_screen.dart';
import '../features/auth/presentation/screens/delete_account_screen.dart';
import '../features/auth/presentation/screens/devices_screen.dart';
import '../features/auth/presentation/screens/forgot_password_screen.dart';
import '../features/auth/presentation/screens/privacy_controls_screen.dart';
import '../features/auth/presentation/screens/reset_password_screen.dart';
import '../features/cart/presentation/screens/cart_screen.dart';
import '../features/categories/presentation/screens/categories_screen.dart';
import '../features/categories/presentation/screens/category_detail_screen.dart';
import '../features/checkout/presentation/screens/checkout_address_screen.dart';
import '../features/checkout/presentation/screens/checkout_payment_screen.dart';
import '../features/checkout/presentation/screens/checkout_review_screen.dart';
import '../features/checkout/presentation/screens/checkout_schedule_screen.dart';
import '../features/coupons/presentation/screens/coupons_screen.dart';
import '../features/favorites/presentation/screens/favorites_screen.dart';
import '../features/help/presentation/screens/help_screen.dart';
import '../features/home/presentation/screens/home_screen.dart';
import '../features/legal/presentation/screens/legal_screen.dart';
import '../features/loyalty/presentation/screens/loyalty_screen.dart';
import '../features/notifications/presentation/screens/notifications_screen.dart';
import '../features/onboarding/presentation/screens/onboarding_screen.dart';
import '../features/orders/presentation/screens/order_detail_screen.dart';
import '../features/orders/presentation/screens/orders_screen.dart';
import '../features/product/presentation/screens/product_screen.dart';
import '../features/profile/presentation/screens/account_screen.dart';
import '../features/profile/presentation/screens/profile_screen.dart';
import '../features/referral/presentation/screens/referral_screen.dart';
import '../features/reviews/presentation/screens/reviews_screen.dart';
import '../features/search/presentation/screens/barcode_scanner_screen.dart';
import '../features/search/presentation/screens/search_screen.dart';
import '../features/settings/presentation/screens/accessibility_settings_screen.dart';
import '../features/settings/presentation/screens/language_settings_screen.dart';
import '../features/settings/presentation/screens/notification_preferences_screen.dart';
import '../features/settings/presentation/screens/privacy_settings_screen.dart';
import '../features/settings/presentation/screens/security_settings_screen.dart';
import '../features/settings/presentation/screens/settings_screen.dart';
import '../features/settings/presentation/screens/theme_settings_screen.dart';
import '../features/shell/presentation/screens/shell_scaffold.dart';
import '../features/splash/presentation/screens/splash_screen.dart';
import '../features/support/presentation/screens/support_assistant_screen.dart';
import '../features/support/presentation/screens/support_screen.dart';
import '../features/support/presentation/screens/support_ticket_detail_screen.dart';
import '../features/tracking/presentation/screens/tracking_screen.dart';
import '../features/wallet/presentation/screens/wallet_screen.dart';
import '../features/about/presentation/screens/about_screen.dart';
import '../features/ai/presentation/screens/ai_hub_screen.dart';
import '../features/ai/presentation/screens/ai_recipes_screen.dart';
import '../features/auth/presentation/screens/auth_email_verify_screen.dart';
import '../features/city/presentation/screens/city_screen.dart';
import 'auth_guard.dart';
import 'route_names.dart';

final _rootNavigatorKey = GlobalKey<NavigatorState>();
final _shellNavigatorHomeKey = GlobalKey<NavigatorState>(debugLabel: 'home');
final _shellNavigatorCategoriesKey =
    GlobalKey<NavigatorState>(debugLabel: 'categories');
final _shellNavigatorSearchKey = GlobalKey<NavigatorState>(debugLabel: 'search');
final _shellNavigatorCartKey = GlobalKey<NavigatorState>(debugLabel: 'cart');
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
        path: RouteNames.onboarding,
        builder: (_, __) => const OnboardingScreen(),
      ),
      GoRoute(
        path: RouteNames.auth,
        builder: (_, __) => const AuthWelcomeScreen(),
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
            path: 'email',
            builder: (_, __) => const AuthEmailScreen(),
          ),
          GoRoute(
            path: 'forgot-password',
            builder: (_, __) => const ForgotPasswordScreen(),
          ),
          GoRoute(
            path: 'reset-password',
            builder: (_, state) => ResetPasswordScreen(
              token: state.uri.queryParameters['token'],
            ),
          ),
          GoRoute(
            path: 'email-verify',
            builder: (_, state) => AuthEmailVerifyScreen(
              email: state.uri.queryParameters['email'],
            ),
          ),
          GoRoute(
            path: 'welcome',
            builder: (_, __) => const AuthWelcomeScreen(),
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
            navigatorKey: _shellNavigatorCategoriesKey,
            routes: [
              GoRoute(
                path: RouteNames.categories,
                builder: (_, __) => const CategoriesScreen(),
                routes: [
                  GoRoute(
                    path: ':categoryId',
                    builder: (_, state) => CategoryDetailScreen(
                      categoryId: state.pathParameters['categoryId']!,
                    ),
                  ),
                ],
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorSearchKey,
            routes: [
              GoRoute(
                path: RouteNames.search,
                builder: (_, __) => const SearchScreen(),
              ),
            ],
          ),
          StatefulShellBranch(
            navigatorKey: _shellNavigatorCartKey,
            routes: [
              GoRoute(
                path: RouteNames.cart,
                builder: (_, __) => const CartScreen(),
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
        path: RouteNames.product,
        builder: (_, state) => ProductScreen(
          productId: state.pathParameters['productId']!,
        ),
      ),
      GoRoute(
        path: RouteNames.checkout,
        redirect: (_, __) => RouteNames.checkoutAddress,
      ),
      GoRoute(
        path: RouteNames.checkoutAddress,
        builder: (_, __) => const CheckoutAddressScreen(),
      ),
      GoRoute(
        path: RouteNames.checkoutSchedule,
        builder: (_, __) => const CheckoutScheduleScreen(),
      ),
      GoRoute(
        path: RouteNames.checkoutPayment,
        builder: (_, __) => const CheckoutPaymentScreen(),
      ),
      GoRoute(
        path: RouteNames.checkoutReview,
        builder: (_, __) => const CheckoutReviewScreen(),
      ),
      GoRoute(
        path: RouteNames.orders,
        builder: (_, __) => const OrdersScreen(),
      ),
      GoRoute(
        path: RouteNames.orderDetail,
        builder: (_, state) => OrderDetailScreen(
          orderId: state.pathParameters['orderId']!,
        ),
      ),
      GoRoute(
        path: RouteNames.orderTrack,
        builder: (_, state) => TrackingScreen(
          orderId: state.pathParameters['orderId']!,
        ),
      ),
      GoRoute(
        path: RouteNames.orderReview,
        builder: (_, state) => ReviewsScreen(
          orderId: state.pathParameters['orderId'],
        ),
      ),
      GoRoute(
        path: RouteNames.favorites,
        builder: (_, __) => const FavoritesScreen(),
      ),
      GoRoute(
        path: RouteNames.wallet,
        builder: (_, __) => const WalletScreen(),
      ),
      GoRoute(
        path: RouteNames.coupons,
        builder: (_, __) => const CouponsScreen(),
      ),
      GoRoute(
        path: RouteNames.loyalty,
        builder: (_, __) => const LoyaltyScreen(),
      ),
      GoRoute(
        path: RouteNames.notifications,
        builder: (_, __) => const NotificationsScreen(),
      ),
      GoRoute(
        path: RouteNames.profile,
        builder: (_, __) => const ProfileScreen(),
      ),
      GoRoute(
        path: RouteNames.addresses,
        builder: (_, __) => const AddressesScreen(),
        routes: [
          GoRoute(
            path: 'add',
            builder: (_, __) => const AddressEditScreen(),
          ),
          GoRoute(
            path: ':addressId/edit',
            builder: (_, state) => AddressEditScreen(
              addressId: state.pathParameters['addressId'],
            ),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.support,
        builder: (_, state) => SupportScreen(orderId: state.uri.queryParameters['orderId']),
        routes: [
          GoRoute(
            path: 'tickets/:ticketId',
            builder: (_, state) => SupportTicketDetailScreen(
              ticketId: state.pathParameters['ticketId']!,
            ),
          ),
          GoRoute(
            path: 'assistant',
            builder: (_, state) => SupportAssistantScreen(
              orderId: state.uri.queryParameters['orderId'],
            ),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.settings,
        builder: (_, __) => const SettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsLanguage,
        builder: (_, __) => const LanguageSettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsTheme,
        builder: (_, __) => const ThemeSettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsA11y,
        builder: (_, __) => const AccessibilitySettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsNotifications,
        builder: (_, __) => const NotificationPreferencesScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsPrivacy,
        builder: (_, __) => const PrivacySettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsSecurity,
        builder: (_, __) => const SecuritySettingsScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsDevices,
        builder: (_, __) => const DevicesScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsDeleteAccount,
        builder: (_, __) => const DeleteAccountScreen(),
      ),
      GoRoute(
        path: RouteNames.settingsPrivacyControls,
        builder: (_, __) => const PrivacyControlsScreen(),
      ),
      GoRoute(
        path: RouteNames.referral,
        builder: (_, __) => const ReferralScreen(),
      ),
      GoRoute(
        path: RouteNames.help,
        builder: (_, __) => const HelpScreen(),
      ),
      GoRoute(
        path: RouteNames.about,
        builder: (_, __) => const AboutScreen(),
      ),
      GoRoute(
        path: RouteNames.legal,
        builder: (_, state) => LegalScreen(
          doc: state.pathParameters['doc']!,
        ),
      ),
      GoRoute(
        path: RouteNames.city,
        builder: (_, __) => const CityScreen(),
      ),
      GoRoute(
        path: RouteNames.ai,
        builder: (_, __) => const AiHubScreen(),
        routes: [
          GoRoute(
            path: 'recipes',
            builder: (_, __) => const AiRecipesScreen(),
          ),
        ],
      ),
      GoRoute(
        path: RouteNames.barcodeScanner,
        builder: (_, __) => const BarcodeScannerScreen(),
      ),
      GoRoute(
        path: RouteNames.campaign,
        builder: (_, state) => HomeScreen(
          campaignSlug: state.pathParameters['slug'],
        ),
      ),
      GoRoute(
        path: RouteNames.promo,
        builder: (_, state) => CouponsScreen(
          promoCode: state.pathParameters['code'],
        ),
      ),
    ],
  );
});
