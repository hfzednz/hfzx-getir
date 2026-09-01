import 'package:equatable/equatable.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../features/addresses/domain/entities/addresses_entity.dart';
import '../../features/checkout/domain/entities/checkout_entity.dart';
import '../utils/money.dart';
import '../validators/address_validator.dart';

/// Aggregated checkout draft for [CheckoutRules.validateDraft].
class CheckoutDraft extends Equatable {
  const CheckoutDraft({
    this.address,
    this.scheduleMode = CheckoutScheduleMode.asap,
    this.scheduledAt,
    this.wantInvoice = false,
    this.invoiceFields,
    this.gift = false,
    this.giftMessage = '',
    this.substitutionPreference = SubstitutionPreference.allow,
    this.outOfStockRule = OutOfStockReplacementRule.similar,
    this.contactless = false,
    this.courierNote = '',
    this.paymentType = 'card',
    this.paymentMethodId,
    this.couponCode,
    this.quote,
  });

  final Address? address;
  final CheckoutScheduleMode scheduleMode;
  final DateTime? scheduledAt;
  final bool wantInvoice;
  final CompanyInvoiceFields? invoiceFields;
  final bool gift;
  final String giftMessage;
  final SubstitutionPreference substitutionPreference;
  final OutOfStockReplacementRule outOfStockRule;
  final bool contactless;
  final String courierNote;
  final String paymentType;
  final String? paymentMethodId;
  final String? couponCode;
  final CheckoutQuote? quote;

  @override
  List<Object?> get props => [
        address,
        scheduleMode,
        scheduledAt,
        wantInvoice,
        invoiceFields,
        gift,
        giftMessage,
        substitutionPreference,
        outOfStockRule,
        contactless,
        courierNote,
        paymentType,
        paymentMethodId,
        couponCode,
        quote,
      ];
}

/// Checkout pre-submit validation helpers.
abstract final class CheckoutRules {
  static const quoteToleranceMinor = 1;

  static Result<Address> validateAddressRequired(Address? address) {
    if (address == null || address.id.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery address is required',
          details: {'field': 'address_id'},
        ),
      );
    }
    return Success(address);
  }

  static Result<Address> validateServiceability({
    required Address? address,
    required bool serviceable,
  }) =>
      AddressValidator.validateForCheckout(address, serviceable: serviceable);

  static Result<void> validateSubstitutionPreferences({
    required SubstitutionPreference substitutionPreference,
    required OutOfStockReplacementRule outOfStockRule,
  }) {
    // Reject is incompatible with auto-replace similar items.
    if (substitutionPreference == SubstitutionPreference.reject &&
        outOfStockRule == OutOfStockReplacementRule.similar) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Cannot auto-replace items when substitutions are rejected',
          details: {
            'substitution_preference': substitutionPreferenceToJson(substitutionPreference),
            'out_of_stock_rule': outOfStockRuleToJson(outOfStockRule),
          },
        ),
      );
    }
    return const Success(null);
  }

  static Result<void> validateDraft({
    required CheckoutDraft draft,
    required bool addressServiceable,
  }) {
    final addressResult = validateServiceability(
      address: draft.address,
      serviceable: addressServiceable,
    );
    if (addressResult.isFailure) return Failure(addressResult.errorOrNull!);

    if (draft.scheduleMode == CheckoutScheduleMode.scheduled) {
      final at = draft.scheduledAt;
      if (at == null || !at.isAfter(DateTime.now())) {
        return const Failure(
          NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Select a future delivery time',
            details: {'field': 'scheduled_at'},
          ),
        );
      }
    }

    if (draft.wantInvoice) {
      final fields = draft.invoiceFields;
      if (fields == null || !fields.isComplete) {
        return const Failure(
          NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Company invoice details are incomplete',
            details: {'field': 'invoice_fields'},
          ),
        );
      }
    }

    if (draft.gift && draft.giftMessage.trim().isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Gift message is required when sending as a gift',
          details: {'field': 'gift_message'},
        ),
      );
    }

    return validateSubstitutionPreferences(
      substitutionPreference: draft.substitutionPreference,
      outOfStockRule: draft.outOfStockRule,
    );
  }

  /// Verifies client-side quote matches server authoritative quote within tolerance.
  static Result<CheckoutQuote> verifyFinalPrice({
    required CheckoutQuote clientQuote,
    required CheckoutQuote serverQuote,
    int toleranceMinor = quoteToleranceMinor,
  }) {
    if (clientQuote.currency != serverQuote.currency) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Price currency mismatch',
          details: {
            'client_currency': clientQuote.currency,
            'server_currency': serverQuote.currency,
          },
        ),
      );
    }

    final delta = (clientQuote.totalMinor - serverQuote.totalMinor).abs();
    if (delta > toleranceMinor) {
      final expected = Money(minorUnits: serverQuote.totalMinor, currency: serverQuote.currency);
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Order total changed to ${expected.format()}',
          details: {
            'client_total_minor': clientQuote.totalMinor,
            'server_total_minor': serverQuote.totalMinor,
            'delta_minor': delta,
          },
        ),
      );
    }

    if (serverQuote.expiresAt != null && serverQuote.expiresAt!.isBefore(DateTime.now())) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Price quote has expired — refresh and try again',
          details: {'quote_id': serverQuote.quoteId},
        ),
      );
    }

    return Success(serverQuote);
  }

  static Result<String> validateVerifiedQuoteId(String? quoteId) {
    final id = quoteId?.trim() ?? '';
    if (id.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Price must be verified before payment',
          details: {'field': 'quote_id'},
        ),
      );
    }
    return Success(id);
  }
}
