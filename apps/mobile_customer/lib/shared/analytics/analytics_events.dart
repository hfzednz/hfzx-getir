import 'package:nexora_core/nexora_core.dart';

/// Typed analytics event names (CONSTITUTION §19 — client UX funnel).
abstract final class AnalyticsEvents {
  static const version = '1';

  // Auth & onboarding
  static const authLoginStarted = 'auth_login_started';
  static const authLoginSucceeded = 'auth_login_succeeded';
  static const authLoginFailed = 'auth_login_failed';
  static const authOtpRequested = 'auth_otp_requested';
  static const authOtpVerified = 'auth_otp_verified';

  // Browse & catalog
  static const homeViewed = 'home_viewed';
  static const searchPerformed = 'search_performed';
  static const productViewed = 'product_viewed';
  static const categoryViewed = 'category_viewed';

  // Cart
  static const cartViewed = 'cart_viewed';
  static const cartItemAdded = 'cart_item_added';
  static const cartItemRemoved = 'cart_item_removed';
  static const cartCouponApplied = 'cart_coupon_applied';
  static const cartCouponRemoved = 'cart_coupon_removed';

  // Checkout funnel
  static const checkoutStarted = 'checkout_started';
  static const checkoutAddressCompleted = 'checkout_address_completed';
  static const checkoutScheduleCompleted = 'checkout_schedule_completed';
  static const checkoutPaymentMethodSelected = 'checkout_payment_method_selected';
  static const checkoutReviewViewed = 'checkout_review_viewed';
  static const checkoutPaymentSubmitted = 'checkout_payment_submitted';
  static const checkoutPaymentSucceeded = 'checkout_payment_succeeded';
  static const checkoutPaymentFailed = 'checkout_payment_failed';
  static const checkoutAbandoned = 'checkout_abandoned';

  // Orders
  static const orderPlaced = 'order_placed';
  static const orderTracked = 'order_tracked';
  static const orderCancelled = 'order_cancelled';
  static const orderReorderTapped = 'order_reorder_tapped';

  // Wallet & promotions
  static const walletTopUpSucceeded = 'wallet_top_up_succeeded';
  static const walletApplyTapped = 'wallet_apply_tapped';
  static const referralShared = 'referral_shared';
  static const reviewSubmitted = 'review_submitted';
  static const supportTicketCreated = 'support_ticket_created';
}

/// Thin wrapper over [AnalyticsGateway] with typed helpers and PII-safe props.
class AnalyticsTracker {
  AnalyticsTracker({
    required AnalyticsGateway gateway,
    this.cityIdProvider,
    this.sessionIdProvider,
    this.appVersionProvider,
  }) : _gateway = gateway;

  final AnalyticsGateway _gateway;
  final String? Function()? cityIdProvider;
  final String? Function()? sessionIdProvider;
  final String? Function()? appVersionProvider;

  Future<void> trackRaw({
    required String eventName,
    Map<String, Object?> props = const {},
    String eventVersion = AnalyticsEvents.version,
  }) =>
      _gateway.track(
        eventName: eventName,
        eventVersion: eventVersion,
        props: {
          if (appVersionProvider?.call() != null) 'app_version': appVersionProvider!(),
          ...props,
        },
        cityId: cityIdProvider?.call(),
        sessionId: sessionIdProvider?.call(),
      );

  Future<void> trackProductViewed({
    required String productId,
    String? categoryId,
    int? priceMinor,
    String? currency,
  }) =>
      trackRaw(
        eventName: AnalyticsEvents.productViewed,
        props: {
          'product_id': productId,
          if (categoryId != null) 'category_id': categoryId,
          if (priceMinor != null) 'price_minor': priceMinor,
          if (currency != null) 'currency': currency,
        },
      );

  Future<void> trackCartItemAdded({
    required String productId,
    required int quantity,
    required int unitPriceMinor,
    String currency = 'TRY',
  }) =>
      trackRaw(
        eventName: AnalyticsEvents.cartItemAdded,
        props: {
          'product_id': productId,
          'quantity': quantity,
          'unit_price_minor': unitPriceMinor,
          'currency': currency,
        },
      );

  Future<void> trackCheckoutStarted({required int itemCount, required int subtotalMinor}) =>
      trackRaw(
        eventName: AnalyticsEvents.checkoutStarted,
        props: {'item_count': itemCount, 'subtotal_minor': subtotalMinor},
      );

  Future<void> trackCheckoutPaymentSubmitted({
    required String paymentMethod,
    required int totalMinor,
    String? quoteId,
  }) =>
      trackRaw(
        eventName: AnalyticsEvents.checkoutPaymentSubmitted,
        props: {
          'payment_method': paymentMethod,
          'total_minor': totalMinor,
          if (quoteId != null) 'quote_id': quoteId,
        },
      );

  Future<void> trackOrderPlaced({
    required String orderId,
    required int totalMinor,
    String currency = 'TRY',
  }) =>
      trackRaw(
        eventName: AnalyticsEvents.orderPlaced,
        props: {
          'order_id': orderId,
          'total_minor': totalMinor,
          'currency': currency,
        },
      );

  Future<void> trackError({
    required String eventName,
    required String errorCode,
    String? surface,
  }) =>
      trackRaw(
        eventName: eventName,
        props: {
          'error_code': errorCode,
          if (surface != null) 'surface': surface,
        },
      );

  Future<void> flush() => _gateway.flush();
}
