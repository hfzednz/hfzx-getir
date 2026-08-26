import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../../shared/business_rules/checkout_rules.dart';
import '../../../../shared/business_rules/payment_rules.dart';
import '../../../addresses/domain/entities/addresses_entity.dart';
import '../../../addresses/presentation/providers/addresses_providers.dart';
import '../../domain/entities/checkout_entity.dart';
import '../../domain/checkout_bff_defaults.dart';
import 'checkout_providers.dart'
    show
        confirmCheckoutUseCaseProvider,
        getCheckoutQuoteUseCaseProvider,
        paymentMethodsRepositoryProvider,
        verifyCheckoutQuoteUseCaseProvider;

class CheckoutState {
  const CheckoutState({
    this.addressId,
    this.paymentMethodId,
    this.paymentType = 'card',
    this.scheduledAt,
    this.contactless = false,
    this.courierNote,
    this.gift = false,
    this.couponCode,
    this.quote,
    this.isLoading = false,
    this.errorMessage,
    this.placedOrderId,
    this.substitutionPreference = SubstitutionPreference.allow,
    this.outOfStockRule = OutOfStockReplacementRule.similar,
    this.wantInvoice = false,
    this.invoiceFields,
    this.giftMessage,
    this.installmentCount,
    this.giftCardCode,
    this.walletAmountMinor,
    this.lastPaymentSessionId,
    this.verifiedQuoteId,
  });

  final String? addressId;
  final String? paymentMethodId;
  final String paymentType;
  final DateTime? scheduledAt;
  final bool contactless;
  final String? courierNote;
  final bool gift;
  final String? couponCode;
  final CheckoutQuote? quote;
  final bool isLoading;
  final String? errorMessage;
  final String? placedOrderId;
  final SubstitutionPreference substitutionPreference;
  final OutOfStockReplacementRule outOfStockRule;
  final bool wantInvoice;
  final CompanyInvoiceFields? invoiceFields;
  final String? giftMessage;
  final int? installmentCount;
  final String? giftCardCode;
  final int? walletAmountMinor;
  final String? lastPaymentSessionId;
  final String? verifiedQuoteId;

  CheckoutState copyWith({
    String? addressId,
    String? paymentMethodId,
    String? paymentType,
    DateTime? scheduledAt,
    bool clearScheduledAt = false,
    bool? contactless,
    String? courierNote,
    bool clearCourierNote = false,
    bool? gift,
    String? couponCode,
    bool clearCouponCode = false,
    CheckoutQuote? quote,
    bool clearQuote = false,
    bool? isLoading,
    String? errorMessage,
    bool clearError = false,
    String? placedOrderId,
    SubstitutionPreference? substitutionPreference,
    OutOfStockReplacementRule? outOfStockRule,
    bool? wantInvoice,
    CompanyInvoiceFields? invoiceFields,
    bool clearInvoiceFields = false,
    String? giftMessage,
    bool clearGiftMessage = false,
    int? installmentCount,
    bool clearInstallmentCount = false,
    String? giftCardCode,
    bool clearGiftCardCode = false,
    int? walletAmountMinor,
    bool clearWalletAmount = false,
    String? lastPaymentSessionId,
    bool clearLastPaymentSessionId = false,
    String? verifiedQuoteId,
    bool clearVerifiedQuoteId = false,
  }) {
    return CheckoutState(
      addressId: addressId ?? this.addressId,
      paymentMethodId: paymentMethodId ?? this.paymentMethodId,
      paymentType: paymentType ?? this.paymentType,
      scheduledAt: clearScheduledAt ? null : (scheduledAt ?? this.scheduledAt),
      contactless: contactless ?? this.contactless,
      courierNote: clearCourierNote ? null : (courierNote ?? this.courierNote),
      gift: gift ?? this.gift,
      couponCode: clearCouponCode ? null : (couponCode ?? this.couponCode),
      quote: clearQuote ? null : (quote ?? this.quote),
      isLoading: isLoading ?? this.isLoading,
      errorMessage: clearError ? null : (errorMessage ?? this.errorMessage),
      placedOrderId: placedOrderId ?? this.placedOrderId,
      substitutionPreference:
          substitutionPreference ?? this.substitutionPreference,
      outOfStockRule: outOfStockRule ?? this.outOfStockRule,
      wantInvoice: wantInvoice ?? this.wantInvoice,
      invoiceFields:
          clearInvoiceFields ? null : (invoiceFields ?? this.invoiceFields),
      giftMessage: clearGiftMessage ? null : (giftMessage ?? this.giftMessage),
      installmentCount: clearInstallmentCount
          ? null
          : (installmentCount ?? this.installmentCount),
      giftCardCode:
          clearGiftCardCode ? null : (giftCardCode ?? this.giftCardCode),
      walletAmountMinor: clearWalletAmount
          ? null
          : (walletAmountMinor ?? this.walletAmountMinor),
      lastPaymentSessionId: clearLastPaymentSessionId
          ? null
          : (lastPaymentSessionId ?? this.lastPaymentSessionId),
      verifiedQuoteId: clearVerifiedQuoteId
          ? null
          : (verifiedQuoteId ?? this.verifiedQuoteId),
    );
  }
}

class CheckoutController extends Notifier<CheckoutState> {
  PaymentAttemptState _paymentAttempt = const PaymentAttemptState();
  NexoraException? _lastPaymentError;

  @override
  CheckoutState build() => const CheckoutState();

  void setAddress(String addressId) {
    state = state.copyWith(addressId: addressId, clearError: true);
  }

  void setPayment({required String paymentType, String? paymentMethodId}) {
    final methodId = paymentMethodId ??
        (paymentType == 'card' ? null : paymentType);
    state = state.copyWith(
      paymentType: paymentType,
      paymentMethodId: methodId,
      clearInstallmentCount: paymentType != 'card',
      clearError: true,
    );
  }

  void setSchedule({DateTime? scheduledAt, bool asap = false}) {
    state = state.copyWith(
      scheduledAt: scheduledAt,
      clearScheduledAt: asap || scheduledAt == null,
      clearError: true,
    );
  }

  void setNotes({
    bool? contactless,
    String? courierNote,
    bool? gift,
    bool clearCourierNote = false,
  }) {
    state = state.copyWith(
      contactless: contactless,
      courierNote: courierNote,
      clearCourierNote: clearCourierNote,
      gift: gift,
      clearGiftMessage: gift == false,
      clearError: true,
    );
  }

  void setSubstitution({
    SubstitutionPreference? preference,
    OutOfStockReplacementRule? outOfStockRule,
  }) {
    var rule = outOfStockRule ?? state.outOfStockRule;
    final pref = preference ?? state.substitutionPreference;
    // Reject + similar is invalid — fall back to refund.
    if (pref == SubstitutionPreference.reject &&
        rule == OutOfStockReplacementRule.similar) {
      rule = OutOfStockReplacementRule.refund;
    }
    state = state.copyWith(
      substitutionPreference: pref,
      outOfStockRule: rule,
      clearError: true,
    );
  }

  void setInvoice({
    bool? wantInvoice,
    CompanyInvoiceFields? fields,
    bool clearFields = false,
  }) {
    state = state.copyWith(
      wantInvoice: wantInvoice,
      invoiceFields: fields,
      clearInvoiceFields: clearFields || wantInvoice == false,
      clearError: true,
    );
  }

  void setGiftMessage(String? message) {
    final trimmed = message?.trim();
    state = state.copyWith(
      giftMessage: trimmed,
      clearGiftMessage: trimmed == null || trimmed.isEmpty,
      clearError: true,
    );
  }

  void setInstallments(int? count) {
    state = state.copyWith(
      installmentCount: count,
      clearInstallmentCount: count == null,
      clearError: true,
    );
  }

  void setGiftCard(String? code) {
    final trimmed = code?.trim();
    state = state.copyWith(
      giftCardCode: trimmed,
      clearGiftCardCode: trimmed == null || trimmed.isEmpty,
      clearError: true,
    );
  }

  void setWalletAmount(int? amountMinor) {
    state = state.copyWith(
      walletAmountMinor: amountMinor,
      clearWalletAmount: amountMinor == null || amountMinor <= 0,
      clearError: true,
    );
  }

  Future<void> applyCoupon(String? code) async {
    final trimmed = code?.trim();
    state = state.copyWith(
      couponCode: trimmed,
      clearCouponCode: trimmed == null || trimmed.isEmpty,
      clearError: true,
    );
    await refreshQuote();
  }

    Map<String, dynamic> buildCheckoutBody() {
    final s = state;
    return {
      'cartId': CheckoutBffDefaults.cartId,
      'principalId': CheckoutBffDefaults.principalId,
      if (s.addressId != null) 'address_id': s.addressId,
      'payment': {
        'type': s.paymentType,
        if (s.paymentMethodId != null) 'payment_method_id': s.paymentMethodId,
        if (s.installmentCount != null) 'installment_count': s.installmentCount,
        if (s.giftCardCode != null && s.giftCardCode!.trim().isNotEmpty)
          'gift_card_code': s.giftCardCode!.trim(),
        if (s.walletAmountMinor != null) 'wallet_amount_minor': s.walletAmountMinor,
      },
      if (s.scheduledAt != null)
        'scheduled_at': s.scheduledAt!.toUtc().toIso8601String(),
      'contactless': s.contactless,
      if (s.courierNote != null && s.courierNote!.trim().isNotEmpty)
        'courier_note': s.courierNote!.trim(),
      'gift': s.gift,
      if (s.gift && s.giftMessage != null && s.giftMessage!.trim().isNotEmpty)
        'gift_message': s.giftMessage!.trim(),
      if (s.couponCode != null && s.couponCode!.trim().isNotEmpty)
        'coupon_code': s.couponCode!.trim(),
      'substitution_preference':
          substitutionPreferenceToJson(s.substitutionPreference),
      'out_of_stock_rule': outOfStockRuleToJson(s.outOfStockRule),
      'want_invoice': s.wantInvoice,
      if (s.wantInvoice && s.invoiceFields != null)
        'invoice_fields': s.invoiceFields!.toJson(),
      if (s.giftCardCode != null && s.giftCardCode!.trim().isNotEmpty)
        'gift_card_code': s.giftCardCode!.trim(),
      if (s.walletAmountMinor != null) 'wallet_amount_minor': s.walletAmountMinor,
      if (s.installmentCount != null) 'installment_count': s.installmentCount,
      if (s.verifiedQuoteId != null) 'quote_id': s.verifiedQuoteId,
      if (s.verifiedQuoteId == null && s.quote?.quoteId != null)
        'quote_id': s.quote!.quoteId,
    };
  }

  Future<void> refreshQuote() async {
    if (state.addressId == null) return;

    state = state.copyWith(isLoading: true, clearError: true, clearVerifiedQuoteId: true);
    final result = await ref.read(getCheckoutQuoteUseCaseProvider).call(
          body: buildCheckoutBody(),
        );
    result.fold(
      onSuccess: (quote) {
        state = state.copyWith(quote: quote, isLoading: false);
      },
      onFailure: (error) {
        state = state.copyWith(
          isLoading: false,
          errorMessage: error.message,
          clearQuote: true,
        );
      },
    );
  }

  CheckoutDraft _buildDraft() {
    final s = state;
    final addresses = ref.read(addressesListProvider).maybeWhen(
          data: (list) => list,
          orElse: () => const <Address>[],
        );
    Address? address;
    if (s.addressId != null) {
      for (final a in addresses) {
        if (a.id == s.addressId) {
          address = a;
          break;
        }
      }
    }
    return CheckoutDraft(
      address: address,
      scheduleMode: s.scheduledAt == null
          ? CheckoutScheduleMode.asap
          : CheckoutScheduleMode.scheduled,
      scheduledAt: s.scheduledAt,
      wantInvoice: s.wantInvoice,
      invoiceFields: s.invoiceFields,
      gift: s.gift,
      giftMessage: s.giftMessage ?? '',
      substitutionPreference: s.substitutionPreference,
      outOfStockRule: s.outOfStockRule,
      contactless: s.contactless,
      courierNote: s.courierNote ?? '',
      paymentType: s.paymentType,
      paymentMethodId: s.paymentMethodId,
      couponCode: s.couponCode,
      quote: s.quote,
    );
  }

  Future<bool> placeOrder({String? idempotencyKey}) async {
    if (state.addressId == null) {
      state = state.copyWith(errorMessage: 'Select a delivery address to continue.');
      return false;
    }
    if (state.paymentMethodId == null && state.paymentType == 'card') {
      state = state.copyWith(errorMessage: 'Select a payment method to continue.');
      return false;
    }
    if (state.paymentType == 'gift_card' &&
        (state.giftCardCode == null || state.giftCardCode!.trim().isEmpty)) {
      state = state.copyWith(errorMessage: 'Enter a gift card code to continue.');
      return false;
    }

    final draft = _buildDraft();
    final draftValidation = CheckoutRules.validateDraft(
      draft: draft,
      addressServiceable: draft.address?.serviceable ?? true,
    );
    if (draftValidation.isFailure) {
      state = state.copyWith(
        errorMessage: draftValidation.errorOrNull!.message,
      );
      return false;
    }

    state = state.copyWith(isLoading: true, clearError: true);

    CheckoutQuote? verifiedServerQuote;
    if (state.quote != null) {
      final quoteId = state.quote!.quoteId?.trim() ?? '';
      if (quoteId.isEmpty) {
        state = state.copyWith(
          isLoading: false,
          errorMessage: 'Price must be verified before payment',
        );
        return false;
      }

      final verifyResult = await ref.read(verifyCheckoutQuoteUseCaseProvider).call(
            quoteId: quoteId,
            body: buildCheckoutBody(),
          );

      final verifiedOk = verifyResult.fold(
        onSuccess: (session) {
          final serverQuote = session.quote;
          if (serverQuote == null) {
            state = state.copyWith(
              isLoading: false,
              errorMessage: 'Price must be verified before payment',
            );
            return false;
          }

          final priceCheck = CheckoutRules.verifyFinalPrice(
            clientQuote: state.quote!,
            serverQuote: serverQuote,
          );
          if (priceCheck.isFailure) {
            state = state.copyWith(
              isLoading: false,
              errorMessage: priceCheck.errorOrNull!.message,
              quote: serverQuote,
              clearVerifiedQuoteId: true,
            );
            return false;
          }

          verifiedServerQuote = serverQuote;
          final verifiedId = serverQuote.quoteId ?? quoteId;
          final idCheck = CheckoutRules.validateVerifiedQuoteId(verifiedId);
          if (idCheck.isFailure) {
            state = state.copyWith(
              isLoading: false,
              errorMessage: idCheck.errorOrNull!.message,
            );
            return false;
          }

          state = state.copyWith(
            quote: serverQuote,
            verifiedQuoteId: idCheck.valueOrNull,
          );
          return true;
        },
        onFailure: (error) {
          state = state.copyWith(isLoading: false, errorMessage: error.message);
          return false;
        },
      );

      if (!verifiedOk) return false;
    } else {
      final idCheck = CheckoutRules.validateVerifiedQuoteId(state.verifiedQuoteId);
      if (idCheck.isFailure) {
        state = state.copyWith(
          isLoading: false,
          errorMessage: idCheck.errorOrNull!.message,
        );
        return false;
      }
    }

    final key = PaymentRules.ensureIdempotencyKey(idempotencyKey);
    final now = DateTime.now();
    _paymentAttempt = _paymentAttempt.copyWith(
      status: PaymentAttemptStatus.inProgress,
      idempotencyKey: key,
      lastAttemptAt: now,
    );

    try {
      unawaited(
        ref.read(analyticsTrackerProvider).trackCheckoutPaymentSubmitted(
              paymentMethod: state.paymentType,
              totalMinor: (verifiedServerQuote ?? state.quote)?.totalMinor ?? 0,
              quoteId: state.verifiedQuoteId ?? state.quote?.quoteId,
            ),
      );
    } catch (_) {
      // Analytics must not block checkout.
    }

    final result = await ref.read(confirmCheckoutUseCaseProvider).call(
          body: buildCheckoutBody(),
          idempotencyKey: key,
        );

    return result.fold(
      onSuccess: (session) {
        final status = session.status.toLowerCase();
        final paymentFailed = status.contains('fail') ||
            status == 'payment_failed' ||
            status == 'requires_payment';

        if (paymentFailed) {
          _lastPaymentError = NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Payment failed — you can retry',
            details: {'session_id': session.id},
          );
          _paymentAttempt = _paymentAttempt.copyWith(
            status: PaymentAttemptStatus.failed,
            paymentIntentId: session.paymentIntentId,
            retryCount: _paymentAttempt.retryCount + 1,
            lastAttemptAt: DateTime.now(),
          );
          state = state.copyWith(
            isLoading: false,
            lastPaymentSessionId: session.id,
            errorMessage: 'Payment failed — you can retry',
            quote: session.quote ?? state.quote,
          );
          try {
            unawaited(
              ref.read(analyticsTrackerProvider).trackRaw(
                    eventName: AnalyticsEvents.checkoutPaymentFailed,
                    props: {
                      'session_id': session.id,
                      'payment_method': state.paymentType,
                    },
                  ),
            );
          } catch (_) {}
          return false;
        }

        final orderId = session.orderId ?? session.id;
        _paymentAttempt = _paymentAttempt.copyWith(
          status: PaymentAttemptStatus.succeeded,
          paymentIntentId: session.paymentIntentId,
          lastAttemptAt: DateTime.now(),
        );
        _lastPaymentError = null;
        state = state.copyWith(
          isLoading: false,
          placedOrderId: orderId,
          quote: session.quote ?? state.quote,
          clearLastPaymentSessionId: true,
        );
        try {
          final total = (session.quote ?? state.quote)?.totalMinor ?? 0;
          final currency = (session.quote ?? state.quote)?.currency ?? 'TRY';
          unawaited(
            ref.read(analyticsTrackerProvider).trackOrderPlaced(
                  orderId: orderId,
                  totalMinor: total,
                  currency: currency,
                ),
          );
          unawaited(
            ref.read(analyticsTrackerProvider).trackRaw(
                  eventName: AnalyticsEvents.checkoutPaymentSucceeded,
                  props: {
                    'order_id': orderId,
                    'payment_method': state.paymentType,
                  },
                ),
          );
        } catch (_) {}
        return true;
      },
      onFailure: (error) {
        _lastPaymentError = error;
        _paymentAttempt = _paymentAttempt.copyWith(
          status: PaymentAttemptStatus.failed,
          retryCount: _paymentAttempt.retryCount + 1,
          lastAttemptAt: DateTime.now(),
        );
        final details = error.details;
        String? sessionIdFromDetails;
        if (details is Map) {
          sessionIdFromDetails = details['session_id']?.toString() ??
              details['payment_session_id']?.toString();
        }
        state = state.copyWith(
          isLoading: false,
          errorMessage: error.message,
          lastPaymentSessionId:
              sessionIdFromDetails ?? state.lastPaymentSessionId,
        );
        try {
          unawaited(
            ref.read(analyticsTrackerProvider).trackRaw(
                  eventName: AnalyticsEvents.checkoutPaymentFailed,
                  props: {
                    'payment_method': state.paymentType,
                    'error_code': error.code.code,
                  },
                ),
          );
        } catch (_) {}
        return false;
      },
    );
  }

  Future<bool> retryPayment(String sessionId) async {
    final eligibility = PaymentRules.validateRetryEligibility(
      state: _paymentAttempt,
      lastError: _lastPaymentError,
    );
    if (eligibility.isFailure) {
      state = state.copyWith(
        errorMessage: eligibility.errorOrNull!.message,
      );
      return false;
    }

    state = state.copyWith(isLoading: true, clearError: true);
    _paymentAttempt = _paymentAttempt.copyWith(
      status: PaymentAttemptStatus.inProgress,
      lastAttemptAt: DateTime.now(),
    );

    final result = await ref
        .read(paymentMethodsRepositoryProvider)
        .retryPayment(sessionId: sessionId);

    return result.fold(
      onSuccess: (session) {
        final status = session.status.toLowerCase();
        final paymentFailed =
            status.contains('fail') || status == 'payment_failed';

        if (paymentFailed) {
          _lastPaymentError = NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Payment failed — you can retry',
            details: {'session_id': session.id},
          );
          _paymentAttempt = _paymentAttempt.copyWith(
            status: PaymentAttemptStatus.failed,
            paymentIntentId: session.paymentIntentId,
            retryCount: _paymentAttempt.retryCount + 1,
            lastAttemptAt: DateTime.now(),
          );
          state = state.copyWith(
            isLoading: false,
            lastPaymentSessionId: session.id,
            errorMessage: 'Payment failed — you can retry',
          );
          return false;
        }

        final orderId = session.orderId ?? session.id;
        _paymentAttempt = _paymentAttempt.copyWith(
          status: PaymentAttemptStatus.succeeded,
          paymentIntentId: session.paymentIntentId,
          lastAttemptAt: DateTime.now(),
        );
        _lastPaymentError = null;
        state = state.copyWith(
          isLoading: false,
          placedOrderId: orderId,
          quote: session.quote ?? state.quote,
          clearLastPaymentSessionId: true,
        );
        try {
          unawaited(
            ref.read(analyticsTrackerProvider).trackOrderPlaced(
                  orderId: orderId,
                  totalMinor:
                      (session.quote ?? state.quote)?.totalMinor ?? 0,
                  currency: (session.quote ?? state.quote)?.currency ?? 'TRY',
                ),
          );
        } catch (_) {}
        return true;
      },
      onFailure: (error) {
        _lastPaymentError = error;
        _paymentAttempt = _paymentAttempt.copyWith(
          status: PaymentAttemptStatus.failed,
          retryCount: _paymentAttempt.retryCount + 1,
          lastAttemptAt: DateTime.now(),
        );
        state = state.copyWith(
          isLoading: false,
          errorMessage: error.message,
          lastPaymentSessionId: sessionId,
        );
        return false;
      },
    );
  }
}
