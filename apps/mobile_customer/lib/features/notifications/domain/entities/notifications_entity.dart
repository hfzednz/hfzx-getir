import 'package:equatable/equatable.dart';

enum NotificationType {
  transactional,
  promo,
  delivery,
  priceDrop,
  restock,
}

NotificationType notificationTypeFromJson(String? raw) {
  switch (raw?.toLowerCase()) {
    case 'transactional':
      return NotificationType.transactional;
    case 'promo':
    case 'promotional':
      return NotificationType.promo;
    case 'delivery':
      return NotificationType.delivery;
    case 'price_drop':
    case 'pricedrop':
      return NotificationType.priceDrop;
    case 'restock':
      return NotificationType.restock;
    default:
      return NotificationType.transactional;
  }
}

/// Primary notification entity (alias: [NotificationItem]).
class AppNotification extends Equatable {
  const AppNotification({
    required this.id,
    this.type = NotificationType.transactional,
    this.title = '',
    this.body = '',
    this.read = false,
    this.deepLink,
    this.createdAt,
    this.payload = const {},
  });

  final String id;
  final NotificationType type;
  final String title;
  final String body;
  final bool read;
  final String? deepLink;
  final DateTime? createdAt;
  final Map<String, dynamic> payload;

  factory AppNotification.fromJson(Map<String, dynamic> json) {
    final payload = Map<String, dynamic>.from(json);
    return AppNotification(
      id: json['id']?.toString() ?? '',
      type: notificationTypeFromJson(json['type']?.toString()),
      title: json['title']?.toString() ?? json['name']?.toString() ?? '',
      body: json['body']?.toString() ??
          json['message']?.toString() ??
          json['subtitle']?.toString() ??
          '',
      read: json['read'] == true || json['is_read'] == true,
      deepLink: json['deep_link']?.toString() ?? json['deepLink']?.toString(),
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'].toString())
          : json['createdAt'] != null
              ? DateTime.tryParse(json['createdAt'].toString())
              : null,
      payload: payload,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type.name,
        'title': title,
        'body': body,
        'read': read,
        if (deepLink != null) 'deep_link': deepLink,
        if (createdAt != null) 'created_at': createdAt!.toIso8601String(),
        ...payload,
      };

  AppNotification copyWith({bool? read}) => AppNotification(
        id: id,
        type: type,
        title: title,
        body: body,
        read: read ?? this.read,
        deepLink: deepLink,
        createdAt: createdAt,
        payload: payload,
      );

  @override
  List<Object?> get props => [id, type, title, body, read, deepLink, createdAt];
}

typedef NotificationItem = AppNotification;

class NotificationPreferences extends Equatable {
  const NotificationPreferences({
    this.transactional = true,
    this.promo = true,
    this.delivery = true,
    this.priceDrop = false,
    this.restock = false,
    this.pushEnabled = true,
    this.emailEnabled = false,
  });

  final bool transactional;
  final bool promo;
  final bool delivery;
  final bool priceDrop;
  final bool restock;
  final bool pushEnabled;
  final bool emailEnabled;

  factory NotificationPreferences.fromJson(Map<String, dynamic> json) =>
      NotificationPreferences(
        transactional: json['transactional'] != false,
        promo: json['promo'] != false,
        delivery: json['delivery'] != false,
        priceDrop: json['price_drop'] == true || json['priceDrop'] == true,
        restock: json['restock'] == true,
        pushEnabled: json['push_enabled'] != false,
        emailEnabled: json['email_enabled'] == true,
      );

  Map<String, dynamic> toJson() => {
        'transactional': transactional,
        'promo': promo,
        'delivery': delivery,
        'price_drop': priceDrop,
        'restock': restock,
        'push_enabled': pushEnabled,
        'email_enabled': emailEnabled,
      };

  NotificationPreferences copyWith({
    bool? transactional,
    bool? promo,
    bool? delivery,
    bool? priceDrop,
    bool? restock,
    bool? pushEnabled,
    bool? emailEnabled,
  }) =>
      NotificationPreferences(
        transactional: transactional ?? this.transactional,
        promo: promo ?? this.promo,
        delivery: delivery ?? this.delivery,
        priceDrop: priceDrop ?? this.priceDrop,
        restock: restock ?? this.restock,
        pushEnabled: pushEnabled ?? this.pushEnabled,
        emailEnabled: emailEnabled ?? this.emailEnabled,
      );

  @override
  List<Object?> get props =>
      [transactional, promo, delivery, priceDrop, restock, pushEnabled, emailEnabled];
}
