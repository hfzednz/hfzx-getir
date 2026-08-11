import 'package:equatable/equatable.dart';

enum SubstitutionPreference { allow, contact, reject }

enum OutOfStockReplacementRule { similar, refund, cancel }

SubstitutionPreference substitutionPreferenceFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'contact':
      return SubstitutionPreference.contact;
    case 'reject':
      return SubstitutionPreference.reject;
    default:
      return SubstitutionPreference.allow;
  }
}

String substitutionPreferenceToJson(SubstitutionPreference pref) => switch (pref) {
      SubstitutionPreference.allow => 'allow',
      SubstitutionPreference.contact => 'contact',
      SubstitutionPreference.reject => 'reject',
    };

OutOfStockReplacementRule outOfStockRuleFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'refund':
      return OutOfStockReplacementRule.refund;
    case 'cancel':
      return OutOfStockReplacementRule.cancel;
    default:
      return OutOfStockReplacementRule.similar;
  }
}

String outOfStockRuleToJson(OutOfStockReplacementRule rule) => switch (rule) {
      OutOfStockReplacementRule.similar => 'similar',
      OutOfStockReplacementRule.refund => 'refund',
      OutOfStockReplacementRule.cancel => 'cancel',
    };

class CompanyInvoiceFields extends Equatable {
  const CompanyInvoiceFields({
    this.companyName,
    this.taxId,
    this.taxOffice,
    this.address,
  });

  final String? companyName;
  final String? taxId;
  final String? taxOffice;
  final String? address;

  factory CompanyInvoiceFields.fromJson(Map<String, dynamic> json) => CompanyInvoiceFields(
        companyName: json['company_name']?.toString(),
        taxId: json['tax_id']?.toString(),
        taxOffice: json['tax_office']?.toString(),
        address: json['address']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        if (companyName != null) 'company_name': companyName,
        if (taxId != null) 'tax_id': taxId,
        if (taxOffice != null) 'tax_office': taxOffice,
        if (address != null) 'address': address,
      };

  bool get isComplete =>
      companyName != null &&
      companyName!.isNotEmpty &&
      taxId != null &&
      taxId!.isNotEmpty;

  @override
  List<Object?> get props => [companyName, taxId, taxOffice, address];
}

class CheckoutQuote extends Equatable {
  const CheckoutQuote({
    required this.subtotalMinor,
    required this.deliveryFeeMinor,
    required this.discountMinor,
    required this.taxMinor,
    required this.totalMinor,
    this.currency = 'TRY',
    this.quoteId,
    this.expiresAt,
  });

  final int subtotalMinor;
  final int deliveryFeeMinor;
  final int discountMinor;
  final int taxMinor;
  final int totalMinor;
  final String currency;
  final String? quoteId;
  final DateTime? expiresAt;

  factory CheckoutQuote.fromJson(Map<String, dynamic> json) => CheckoutQuote(
        subtotalMinor: (json['subtotal_minor'] as num?)?.toInt() ?? 0,
        deliveryFeeMinor: (json['delivery_fee_minor'] as num?)?.toInt() ?? 0,
        discountMinor: (json['discount_minor'] as num?)?.toInt() ?? 0,
        taxMinor: (json['tax_minor'] as num?)?.toInt() ?? 0,
        totalMinor: (json['total_minor'] as num?)?.toInt() ?? 0,
        currency: json['currency']?.toString() ?? 'TRY',
        quoteId: json['quote_id']?.toString(),
        expiresAt: json['expires_at'] != null
            ? DateTime.tryParse(json['expires_at'].toString())
            : null,
      );

  Map<String, dynamic> toJson() => {
        'subtotal_minor': subtotalMinor,
        'delivery_fee_minor': deliveryFeeMinor,
        'discount_minor': discountMinor,
        'tax_minor': taxMinor,
        'total_minor': totalMinor,
        'currency': currency,
        if (quoteId != null) 'quote_id': quoteId,
        if (expiresAt != null) 'expires_at': expiresAt!.toUtc().toIso8601String(),
      };

  @override
  List<Object?> get props =>
      [subtotalMinor, deliveryFeeMinor, discountMinor, taxMinor, totalMinor, currency, quoteId, expiresAt];
}

enum CheckoutScheduleMode { asap, scheduled }

class SavedPaymentCard extends Equatable {
  const SavedPaymentCard({
    required this.id,
    required this.last4,
    required this.brand,
    this.expiryMonth,
    this.expiryYear,
    this.isDefault = false,
  });

  final String id;
  final String last4;
  final String brand;
  final int? expiryMonth;
  final int? expiryYear;
  final bool isDefault;

  factory SavedPaymentCard.fromJson(Map<String, dynamic> json) => SavedPaymentCard(
        id: json['id']?.toString() ?? '',
        last4: json['last4']?.toString() ?? json['last_4']?.toString() ?? '',
        brand: json['brand']?.toString() ?? '',
        expiryMonth: (json['expiry_month'] as num?)?.toInt(),
        expiryYear: (json['expiry_year'] as num?)?.toInt(),
        isDefault: json['is_default'] == true,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'last4': last4,
        'brand': brand,
        if (expiryMonth != null) 'expiry_month': expiryMonth,
        if (expiryYear != null) 'expiry_year': expiryYear,
        'is_default': isDefault,
      };

  @override
  List<Object?> get props => [id, last4, brand, expiryMonth, expiryYear, isDefault];
}

class CheckoutSession extends Equatable {
  const CheckoutSession({
    required this.id,
    this.orderId,
    this.status = 'pending',
    this.quote,
    this.substitutionPreference = SubstitutionPreference.allow,
    this.outOfStockRule = OutOfStockReplacementRule.similar,
    this.invoiceFields,
    this.paymentIntentId,
  });

  final String id;
  final String? orderId;
  final String status;
  final CheckoutQuote? quote;
  final SubstitutionPreference substitutionPreference;
  final OutOfStockReplacementRule outOfStockRule;
  final CompanyInvoiceFields? invoiceFields;
  final String? paymentIntentId;

  factory CheckoutSession.fromJson(Map<String, dynamic> json) => CheckoutSession(
        id: json['id']?.toString() ?? '',
        orderId: json['order_id']?.toString(),
        status: json['status']?.toString() ?? 'pending',
        quote: json['quote'] != null
            ? CheckoutQuote.fromJson(json['quote'] as Map<String, dynamic>)
            : null,
        substitutionPreference:
            substitutionPreferenceFromJson(json['substitution_preference']?.toString()),
        outOfStockRule: outOfStockRuleFromJson(json['out_of_stock_rule']?.toString()),
        invoiceFields: json['invoice_fields'] != null
            ? CompanyInvoiceFields.fromJson(json['invoice_fields'] as Map<String, dynamic>)
            : null,
        paymentIntentId: json['payment_intent_id']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        if (orderId != null) 'order_id': orderId,
        'status': status,
        if (quote != null) 'quote': quote!.toJson(),
        'substitution_preference': substitutionPreferenceToJson(substitutionPreference),
        'out_of_stock_rule': outOfStockRuleToJson(outOfStockRule),
        if (invoiceFields != null) 'invoice_fields': invoiceFields!.toJson(),
        if (paymentIntentId != null) 'payment_intent_id': paymentIntentId,
      };

  @override
  List<Object?> get props => [
        id,
        orderId,
        status,
        quote,
        substitutionPreference,
        outOfStockRule,
        invoiceFields,
        paymentIntentId,
      ];
}
