import 'package:nexora_core/nexora_core.dart';

/// Courier analytics event names.
abstract final class CourierAnalyticsEvents {
  static const version = '1';

  // Auth & KYC
  static const authOtpRequested = 'courier_auth_otp_requested';
  static const authOtpVerified = 'courier_auth_otp_verified';
  static const authLoginFailed = 'courier_auth_login_failed';
  static const kycUploadStarted = 'courier_kyc_upload_started';
  static const kycUploadSucceeded = 'courier_kyc_upload_succeeded';
  static const kycApproved = 'courier_kyc_approved';

  // Duty
  static const dutyStatusChanged = 'courier_duty_status_changed';
  static const wentOnline = 'courier_went_online';
  static const wentOffline = 'courier_went_offline';
  static const emergencyActivated = 'courier_emergency_activated';

  // Offers
  static const offerReceived = 'courier_offer_received';
  static const offerAccepted = 'courier_offer_accepted';
  static const offerRejected = 'courier_offer_rejected';
  static const offerExpired = 'courier_offer_expired';

  // Delivery workflow
  static const deliveryAssigned = 'courier_delivery_assigned';
  static const deliveryStatusChanged = 'courier_delivery_status_changed';
  static const pickupConfirmed = 'courier_pickup_confirmed';
  static const podSubmitted = 'courier_pod_submitted';
  static const deliveryFailed = 'courier_delivery_failed';
  static const navigationStarted = 'courier_navigation_started';

  // Earnings / shifts
  static const shiftStarted = 'courier_shift_started';
  static const shiftEnded = 'courier_shift_ended';
  static const earningsViewed = 'courier_earnings_viewed';

  // Location / integrity
  static const locationSpoofFlagged = 'courier_location_spoof_flagged';
  static const lowBatteryMode = 'courier_low_battery_mode';
}

/// Thin tracker over [AnalyticsGateway] with courier-safe props.
class CourierAnalyticsTracker {
  CourierAnalyticsTracker({
    required AnalyticsGateway gateway,
    this.cityIdProvider,
    this.courierIdProvider,
    this.sessionIdProvider,
  }) : _gateway = gateway;

  final AnalyticsGateway _gateway;
  final String? Function()? cityIdProvider;
  final String? Function()? courierIdProvider;
  final String? Function()? sessionIdProvider;

  Future<void> track({
    required String eventName,
    Map<String, Object?> props = const {},
    String eventVersion = CourierAnalyticsEvents.version,
  }) =>
      _gateway.track(
        eventName: eventName,
        eventVersion: eventVersion,
        props: {
          if (courierIdProvider?.call() != null)
            'courier_id': courierIdProvider!(),
          ...props,
        },
        cityId: cityIdProvider?.call(),
        sessionId: sessionIdProvider?.call(),
      );

  Future<void> trackDutyChanged({
    required String from,
    required String to,
  }) =>
      track(
        eventName: CourierAnalyticsEvents.dutyStatusChanged,
        props: {'from': from, 'to': to},
      );

  Future<void> trackOfferAccepted({
    required String offerId,
    required String deliveryId,
    int? feeMinor,
    String currency = 'TRY',
  }) =>
      track(
        eventName: CourierAnalyticsEvents.offerAccepted,
        props: {
          'offer_id': offerId,
          'delivery_id': deliveryId,
          if (feeMinor != null) 'fee_minor': feeMinor,
          'currency': currency,
        },
      );

  Future<void> trackPodSubmitted({
    required String deliveryId,
    required bool hasPhoto,
    required bool hasSignature,
  }) =>
      track(
        eventName: CourierAnalyticsEvents.podSubmitted,
        props: {
          'delivery_id': deliveryId,
          'has_photo': hasPhoto,
          'has_signature': hasSignature,
        },
      );

  Future<void> trackEarningsViewed({
    required int totalMinor,
    String currency = 'TRY',
    String period = 'today',
  }) =>
      track(
        eventName: CourierAnalyticsEvents.earningsViewed,
        props: {
          'total_minor': totalMinor,
          'currency': currency,
          'period': period,
        },
      );

  Future<void> flush() => _gateway.flush();
}
