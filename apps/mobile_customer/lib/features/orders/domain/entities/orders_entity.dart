import 'package:equatable/equatable.dart';

import '../../../../shared/business_rules/order_rules.dart';

class OrderLineItem extends Equatable {
  const OrderLineItem({
    required this.id,
    required this.productId,
    required this.name,
    required this.quantity,
    required this.unitPriceMinor,
    this.imageUrl,
    this.sku,
    this.cancelled = false,
  });

  final String id;
  final String productId;
  final String name;
  final int quantity;
  final int unitPriceMinor;
  final String? imageUrl;
  final String? sku;
  final bool cancelled;

  int get lineTotalMinor => unitPriceMinor * quantity;

  factory OrderLineItem.fromJson(Map<String, dynamic> json) => OrderLineItem(
        id: json['id']?.toString() ?? json['line_id']?.toString() ?? '',
        productId: json['product_id']?.toString() ??
            json['productId']?.toString() ??
            json['sku']?.toString() ??
            '',
        name: json['name']?.toString() ?? json['title']?.toString() ?? '',
        quantity: (json['quantity'] as num?)?.toInt() ??
            (json['qty'] as num?)?.toInt() ??
            1,
        unitPriceMinor: (json['unit_price_minor'] as num?)?.toInt() ??
            (json['unitPriceMinor'] as num?)?.toInt() ??
            (json['price_minor'] as num?)?.toInt() ??
            (json['priceMinor'] as num?)?.toInt() ??
            0,
        imageUrl: json['image_url']?.toString() ?? json['imageUrl']?.toString(),
        sku: json['sku']?.toString(),
        cancelled: json['cancelled'] == true || json['is_cancelled'] == true,
      );

  @override
  List<Object?> get props =>
      [id, productId, name, quantity, unitPriceMinor, imageUrl, sku, cancelled];
}

class OrderTotals extends Equatable {
  const OrderTotals({
    this.subtotalMinor = 0,
    this.discountMinor = 0,
    this.deliveryMinor = 0,
    this.serviceMinor = 0,
    this.taxMinor = 0,
    this.totalMinor = 0,
  });

  final int subtotalMinor;
  final int discountMinor;
  final int deliveryMinor;
  final int serviceMinor;
  final int taxMinor;
  final int totalMinor;

  factory OrderTotals.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const OrderTotals();
    return OrderTotals(
      subtotalMinor: (json['subtotal_minor'] as num?)?.toInt() ?? 0,
      discountMinor: (json['discount_minor'] as num?)?.toInt() ?? 0,
      deliveryMinor: (json['delivery_minor'] as num?)?.toInt() ?? 0,
      serviceMinor: (json['service_minor'] as num?)?.toInt() ?? 0,
      taxMinor: (json['tax_minor'] as num?)?.toInt() ?? 0,
      totalMinor: (json['total_minor'] as num?)?.toInt() ??
          (json['grand_total_minor'] as num?)?.toInt() ??
          0,
    );
  }

  @override
  List<Object?> get props =>
      [subtotalMinor, discountMinor, deliveryMinor, serviceMinor, taxMinor, totalMinor];
}

class OrderCourierSummary extends Equatable {
  const OrderCourierSummary({
    this.id,
    this.name,
    this.phone,
    this.vehicle,
    this.rating,
    this.chatUrl,
  });

  final String? id;
  final String? name;
  final String? phone;
  final String? vehicle;
  final double? rating;
  final String? chatUrl;

  factory OrderCourierSummary.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const OrderCourierSummary();
    return OrderCourierSummary(
      id: json['id']?.toString(),
      name: json['name']?.toString(),
      phone: json['phone']?.toString(),
      vehicle: json['vehicle']?.toString(),
      rating: (json['rating'] as num?)?.toDouble(),
      chatUrl: json['chat_url']?.toString() ?? json['chatUrl']?.toString(),
    );
  }

  @override
  List<Object?> get props => [id, name, phone, vehicle, rating, chatUrl];
}

class OrderTimelineEvent extends Equatable {
  const OrderTimelineEvent({
    required this.label,
    this.subtitle,
    this.timestamp,
    this.state = 'upcoming',
  });

  final String label;
  final String? subtitle;
  final DateTime? timestamp;
  final String state;

  factory OrderTimelineEvent.fromJson(Map<String, dynamic> json) =>
      OrderTimelineEvent(
        label: json['label']?.toString() ??
            json['title']?.toString() ??
            '',
        subtitle: json['subtitle']?.toString(),
        timestamp: json['timestamp'] != null
            ? DateTime.tryParse(json['timestamp'].toString())
            : null,
        state: json['state']?.toString() ?? 'upcoming',
      );

  @override
  List<Object?> get props => [label, subtitle, timestamp, state];
}

class OrderCancellationPolicy extends Equatable {
  const OrderCancellationPolicy({
    this.cancellableUntil,
    this.partialCancelAllowed = false,
    this.refundEligible = false,
    this.policyText,
  });

  final DateTime? cancellableUntil;
  final bool partialCancelAllowed;
  final bool refundEligible;
  final String? policyText;

  factory OrderCancellationPolicy.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const OrderCancellationPolicy();
    return OrderCancellationPolicy(
      cancellableUntil: json['cancellable_until'] != null
          ? DateTime.tryParse(json['cancellable_until'].toString())
          : null,
      partialCancelAllowed: json['partial_cancel_allowed'] == true,
      refundEligible: json['refund_eligible'] == true,
      policyText: json['policy_text']?.toString(),
    );
  }

  @override
  List<Object?> get props =>
      [cancellableUntil, partialCancelAllowed, refundEligible, policyText];
}

class Order extends Equatable {
  const Order({
    required this.id,
    this.status = OrderLifecycleStatus.confirmed,
    this.title = '',
    this.items = const [],
    this.totals = const OrderTotals(),
    this.currency = 'TRY',
    this.etaMinutes,
    this.courier,
    this.timeline = const [],
    this.invoiceUrl,
    this.receiptUrl,
    this.proofOfDeliveryUrl,
    this.proofOfDeliveryPhotos = const [],
    this.cancellationPolicy = const OrderCancellationPolicy(),
    this.isFavorite = false,
    this.paymentCaptured = false,
    this.allItemsAvailable = true,
    this.payload = const {},
  });

  final String id;
  final OrderLifecycleStatus status;
  final String title;
  final List<OrderLineItem> items;
  final OrderTotals totals;
  final String currency;
  final int? etaMinutes;
  final OrderCourierSummary? courier;
  final List<OrderTimelineEvent> timeline;
  final String? invoiceUrl;
  final String? receiptUrl;
  final String? proofOfDeliveryUrl;
  final List<String> proofOfDeliveryPhotos;
  final OrderCancellationPolicy cancellationPolicy;
  final bool isFavorite;
  final bool paymentCaptured;
  final bool allItemsAvailable;
  final Map<String, dynamic> payload;

  bool get hasProofOfDelivery =>
      (proofOfDeliveryUrl != null && proofOfDeliveryUrl!.isNotEmpty) ||
      proofOfDeliveryPhotos.isNotEmpty;

  String get statusLabel => status.name;

  String get totalLabel {
    final major = totals.totalMinor / 100;
    return '$currency ${major.toStringAsFixed(2)}';
  }

  bool get canCancel => OrderRules.isCancellable(status);

  bool get canPartialCancel => OrderRules.allowsPartialCancel(status);

  bool get canReorder => OrderRules.isReorderEligible(status);

  factory Order.fromJson(Map<String, dynamic> json) {
    final payload = Map<String, dynamic>.from(json);
    final itemsRaw = json['items'] as List<dynamic>? ??
        json['lines'] as List<dynamic>? ??
        [];
    final timelineRaw = json['timeline'] as List<dynamic>? ??
        json['events'] as List<dynamic>? ??
        [];

    return Order(
      id: json['id']?.toString() ??
          json['orderId']?.toString() ??
          json['order_id']?.toString() ??
          '',
      status: orderLifecycleStatusFromJson(json['status']?.toString()),
      title: json['title']?.toString() ?? json['name']?.toString() ?? '',
      items: [
        for (final e in itemsRaw)
          if (e is Map)
            OrderLineItem.fromJson(Map<String, dynamic>.from(e)),
      ],
      totals: OrderTotals.fromJson(
        json['totals'] as Map<String, dynamic>? ?? json,
      ),
      currency: json['currency']?.toString() ?? 'TRY',
      etaMinutes: (json['eta_minutes'] as num?)?.toInt(),
      courier: OrderCourierSummary.fromJson(
        json['courier'] as Map<String, dynamic>? ??
            json['courier_summary'] as Map<String, dynamic>?,
      ),
      timeline: timelineRaw
          .map((e) => OrderTimelineEvent.fromJson(e as Map<String, dynamic>))
          .toList(),
      invoiceUrl: json['invoice_url']?.toString() ?? json['invoiceUrl']?.toString(),
      receiptUrl: json['receipt_url']?.toString() ?? json['receiptUrl']?.toString(),
      proofOfDeliveryUrl: json['proof_of_delivery_url']?.toString() ??
          json['proofOfDeliveryUrl']?.toString() ??
          json['pod_url']?.toString(),
      proofOfDeliveryPhotos: _parseProofPhotos(json),
      cancellationPolicy: OrderCancellationPolicy.fromJson(
        json['cancellation_policy'] as Map<String, dynamic>?,
      ),
      isFavorite: json['is_favorite'] == true || json['isFavorite'] == true,
      paymentCaptured:
          json['payment_captured'] == true || json['paymentCaptured'] == true,
      allItemsAvailable: json['all_items_available'] != false,
      payload: payload,
    );
  }

  static List<String> _parseProofPhotos(Map<String, dynamic> json) {
    final raw = json['proof_of_delivery_photos'] ??
        json['proofOfDeliveryPhotos'] ??
        json['pod_photos'] ??
        json['pod_images'];
    if (raw is! List) return const [];
    return raw.map((e) => e.toString()).where((u) => u.isNotEmpty).toList();
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'status': status.name,
        'title': title,
        'items': items.map((i) => {
              'id': i.id,
              'product_id': i.productId,
              'name': i.name,
              'quantity': i.quantity,
              'unit_price_minor': i.unitPriceMinor,
            }).toList(),
        'totals': {
          'subtotal_minor': totals.subtotalMinor,
          'discount_minor': totals.discountMinor,
          'delivery_minor': totals.deliveryMinor,
          'service_minor': totals.serviceMinor,
          'tax_minor': totals.taxMinor,
          'total_minor': totals.totalMinor,
        },
        'currency': currency,
        if (etaMinutes != null) 'eta_minutes': etaMinutes,
        if (invoiceUrl != null) 'invoice_url': invoiceUrl,
        if (receiptUrl != null) 'receipt_url': receiptUrl,
        if (proofOfDeliveryUrl != null)
          'proof_of_delivery_url': proofOfDeliveryUrl,
        if (proofOfDeliveryPhotos.isNotEmpty)
          'proof_of_delivery_photos': proofOfDeliveryPhotos,
        'is_favorite': isFavorite,
        'payment_captured': paymentCaptured,
        ...payload,
      };

  @override
  List<Object?> get props => [
        id,
        status,
        title,
        items,
        totals,
        currency,
        etaMinutes,
        courier,
        timeline,
        invoiceUrl,
        receiptUrl,
        proofOfDeliveryUrl,
        proofOfDeliveryPhotos,
        cancellationPolicy,
        isFavorite,
        paymentCaptured,
      ];
}

class OrderDocument extends Equatable {
  const OrderDocument({required this.url, this.mimeType, this.expiresAt});

  final String url;
  final String? mimeType;
  final DateTime? expiresAt;

  factory OrderDocument.fromJson(Map<String, dynamic> json) => OrderDocument(
        url: json['url']?.toString() ?? '',
        mimeType: json['mime_type']?.toString(),
        expiresAt: json['expires_at'] != null
            ? DateTime.tryParse(json['expires_at'].toString())
            : null,
      );

  @override
  List<Object?> get props => [url, mimeType, expiresAt];
}
