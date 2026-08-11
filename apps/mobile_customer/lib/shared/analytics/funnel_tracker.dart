import 'analytics_events.dart';

/// Checkout funnel steps (home → PDP → cart → pay → delivered).
enum CheckoutFunnelStep {
  cartViewed,
  checkoutStarted,
  addressCompleted,
  scheduleCompleted,
  paymentMethodSelected,
  reviewViewed,
  paymentSubmitted,
  paymentSucceeded,
}

extension CheckoutFunnelStepX on CheckoutFunnelStep {
  String get analyticsEvent => switch (this) {
        CheckoutFunnelStep.cartViewed => AnalyticsEvents.cartViewed,
        CheckoutFunnelStep.checkoutStarted => AnalyticsEvents.checkoutStarted,
        CheckoutFunnelStep.addressCompleted => AnalyticsEvents.checkoutAddressCompleted,
        CheckoutFunnelStep.scheduleCompleted => AnalyticsEvents.checkoutScheduleCompleted,
        CheckoutFunnelStep.paymentMethodSelected =>
          AnalyticsEvents.checkoutPaymentMethodSelected,
        CheckoutFunnelStep.reviewViewed => AnalyticsEvents.checkoutReviewViewed,
        CheckoutFunnelStep.paymentSubmitted => AnalyticsEvents.checkoutPaymentSubmitted,
        CheckoutFunnelStep.paymentSucceeded => AnalyticsEvents.checkoutPaymentSucceeded,
      };

  CheckoutFunnelStep? get previous => switch (this) {
        CheckoutFunnelStep.cartViewed => null,
        CheckoutFunnelStep.checkoutStarted => CheckoutFunnelStep.cartViewed,
        CheckoutFunnelStep.addressCompleted => CheckoutFunnelStep.checkoutStarted,
        CheckoutFunnelStep.scheduleCompleted => CheckoutFunnelStep.addressCompleted,
        CheckoutFunnelStep.paymentMethodSelected => CheckoutFunnelStep.scheduleCompleted,
        CheckoutFunnelStep.reviewViewed => CheckoutFunnelStep.paymentMethodSelected,
        CheckoutFunnelStep.paymentSubmitted => CheckoutFunnelStep.reviewViewed,
        CheckoutFunnelStep.paymentSucceeded => CheckoutFunnelStep.paymentSubmitted,
      };
}

/// Tracks sequential checkout funnel progression and drop-off.
class FunnelTracker {
  FunnelTracker(this._tracker);

  final AnalyticsTracker _tracker;

  CheckoutFunnelStep? _lastStep;
  DateTime? _funnelStartedAt;

  CheckoutFunnelStep? get lastStep => _lastStep;

  Future<void> advance(
    CheckoutFunnelStep step, {
    Map<String, Object?> props = const {},
  }) async {
    final expectedPrevious = step.previous;
    if (expectedPrevious != null && _lastStep != null && _lastStep != expectedPrevious) {
      await _tracker.trackRaw(
        eventName: AnalyticsEvents.checkoutAbandoned,
        props: {
          'last_step': _lastStep!.name,
          'attempted_step': step.name,
          ...props,
        },
      );
    }

    if (step == CheckoutFunnelStep.checkoutStarted) {
      _funnelStartedAt = DateTime.now();
    }

    _lastStep = step;

    final enriched = <String, Object?>{
      'funnel_step': step.name,
      if (_funnelStartedAt != null)
        'seconds_since_funnel_start':
            DateTime.now().difference(_funnelStartedAt!).inSeconds,
      ...props,
    };

    await _tracker.trackRaw(eventName: step.analyticsEvent, props: enriched);
  }

  Future<void> completePayment({
    required int totalMinor,
    String? orderId,
    String currency = 'TRY',
  }) async {
    await advance(
      CheckoutFunnelStep.paymentSucceeded,
      props: {
        'total_minor': totalMinor,
        'currency': currency,
        if (orderId != null) 'order_id': orderId,
      },
    );
    _funnelStartedAt = null;
    _lastStep = null;
  }

  Future<void> abandon({String? reason}) async {
    if (_lastStep == null) return;
    await _tracker.trackRaw(
      eventName: AnalyticsEvents.checkoutAbandoned,
      props: {
        'last_step': _lastStep!.name,
        if (reason != null) 'reason': reason,
      },
    );
    _funnelStartedAt = null;
    _lastStep = null;
  }

  void reset() {
    _lastStep = null;
    _funnelStartedAt = null;
  }
}
