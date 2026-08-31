import 'package:equatable/equatable.dart';

class CommunicationPreferences extends Equatable {
  const CommunicationPreferences({
    this.emailMarketing = true,
    this.pushOrderUpdates = true,
    this.pushPromotions = false,
    this.smsAlerts = false,
  });

  final bool emailMarketing;
  final bool pushOrderUpdates;
  final bool pushPromotions;
  final bool smsAlerts;

  factory CommunicationPreferences.fromJson(Map<String, dynamic>? json) {
    if (json == null) return const CommunicationPreferences();
    return CommunicationPreferences(
      emailMarketing: json['email_marketing'] as bool? ?? true,
      pushOrderUpdates: json['push_order_updates'] as bool? ?? true,
      pushPromotions: json['push_promotions'] as bool? ?? false,
      smsAlerts: json['sms_alerts'] as bool? ?? false,
    );
  }

  Map<String, dynamic> toJson() => {
        'email_marketing': emailMarketing,
        'push_order_updates': pushOrderUpdates,
        'push_promotions': pushPromotions,
        'sms_alerts': smsAlerts,
      };

  CommunicationPreferences copyWith({
    bool? emailMarketing,
    bool? pushOrderUpdates,
    bool? pushPromotions,
    bool? smsAlerts,
  }) =>
      CommunicationPreferences(
        emailMarketing: emailMarketing ?? this.emailMarketing,
        pushOrderUpdates: pushOrderUpdates ?? this.pushOrderUpdates,
        pushPromotions: pushPromotions ?? this.pushPromotions,
        smsAlerts: smsAlerts ?? this.smsAlerts,
      );

  @override
  List<Object?> get props => [emailMarketing, pushOrderUpdates, pushPromotions, smsAlerts];
}

class UserProfile extends Equatable {
  const UserProfile({
    required this.id,
    this.firstName = '',
    this.lastName = '',
    this.displayName = '',
    this.email = '',
    this.phone = '',
    this.avatarUrl,
    this.dateOfBirth,
    this.communicationPreferences = const CommunicationPreferences(),
    this.locale = 'en',
    this.timezone,
  });

  final String id;
  final String firstName;
  final String lastName;
  final String displayName;
  final String email;
  final String phone;
  final String? avatarUrl;
  final DateTime? dateOfBirth;
  final CommunicationPreferences communicationPreferences;
  final String locale;
  final String? timezone;

  String get fullName {
    if (displayName.isNotEmpty) return displayName;
    return [firstName, lastName].where((s) => s.isNotEmpty).join(' ');
  }

  factory UserProfile.fromJson(Map<String, dynamic> json) => UserProfile(
        id: json['id']?.toString() ??
            json['customerId']?.toString() ??
            json['customer_id']?.toString() ??
            '',
        firstName: json['first_name']?.toString() ?? json['firstName']?.toString() ?? '',
        lastName: json['last_name']?.toString() ?? json['lastName']?.toString() ?? '',
        displayName: json['display_name']?.toString() ?? json['name']?.toString() ?? '',
        email: json['email']?.toString() ?? '',
        phone: json['phone']?.toString() ?? '',
        avatarUrl: json['avatar_url']?.toString(),
        dateOfBirth:
            json['date_of_birth'] != null ? DateTime.tryParse(json['date_of_birth'].toString()) : null,
        communicationPreferences: CommunicationPreferences.fromJson(
          json['communication_preferences'] as Map<String, dynamic>?,
        ),
        locale: json['locale']?.toString() ?? 'en',
        timezone: json['timezone']?.toString(),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        if (firstName.isNotEmpty) 'first_name': firstName,
        if (lastName.isNotEmpty) 'last_name': lastName,
        if (displayName.isNotEmpty) 'display_name': displayName,
        if (email.isNotEmpty) 'email': email,
        if (phone.isNotEmpty) 'phone': phone,
        if (avatarUrl != null) 'avatar_url': avatarUrl,
        if (dateOfBirth != null) 'date_of_birth': dateOfBirth!.toIso8601String().split('T').first,
        'communication_preferences': communicationPreferences.toJson(),
        'locale': locale,
        if (timezone != null) 'timezone': timezone,
      };

  UserProfile copyWith({
    String? firstName,
    String? lastName,
    String? displayName,
    String? email,
    String? phone,
    String? avatarUrl,
    DateTime? dateOfBirth,
    CommunicationPreferences? communicationPreferences,
    String? locale,
    String? timezone,
  }) =>
      UserProfile(
        id: id,
        firstName: firstName ?? this.firstName,
        lastName: lastName ?? this.lastName,
        displayName: displayName ?? this.displayName,
        email: email ?? this.email,
        phone: phone ?? this.phone,
        avatarUrl: avatarUrl ?? this.avatarUrl,
        dateOfBirth: dateOfBirth ?? this.dateOfBirth,
        communicationPreferences: communicationPreferences ?? this.communicationPreferences,
        locale: locale ?? this.locale,
        timezone: timezone ?? this.timezone,
      );

  @override
  List<Object?> get props => [id, email, phone, displayName, communicationPreferences];
}
