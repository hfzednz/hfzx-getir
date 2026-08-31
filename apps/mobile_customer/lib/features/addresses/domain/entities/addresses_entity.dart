import 'package:equatable/equatable.dart';

enum AddressLabel { home, work, custom }

AddressLabel addressLabelFromJson(String? value) {
  switch (value?.toLowerCase()) {
    case 'work':
      return AddressLabel.work;
    case 'custom':
      return AddressLabel.custom;
    default:
      return AddressLabel.home;
  }
}

String addressLabelToJson(AddressLabel label) => switch (label) {
      AddressLabel.home => 'home',
      AddressLabel.work => 'work',
      AddressLabel.custom => 'custom',
    };

class AddressZoneValidation extends Equatable {
  const AddressZoneValidation({
    required this.serviceable,
    this.cityId,
    this.storeId,
    this.message,
  });

  final bool serviceable;
  final String? cityId;
  final String? storeId;
  final String? message;

  factory AddressZoneValidation.fromJson(Map<String, dynamic> json) =>
      AddressZoneValidation(
        serviceable: json['serviceable'] == true || json['is_serviceable'] == true,
        cityId: json['city_id']?.toString(),
        storeId: json['store_id']?.toString(),
        message: json['message']?.toString(),
      );

  @override
  List<Object?> get props => [serviceable, cityId, storeId, message];
}

class Address extends Equatable {
  const Address({
    required this.id,
    this.title = '',
    this.label = AddressLabel.home,
    this.customLabel = '',
    this.formatted = '',
    this.building = '',
    this.floor = '',
    this.door = '',
    this.deliveryInstructions = '',
    this.lat,
    this.lng,
    this.cityId,
    this.isFavorite = false,
    this.isDefault = false,
    this.serviceable = true,
    this.payload = const {},
    this.recipientName = '',
    this.recipientPhone = '',
  });

  final String id;
  final String title;
  final AddressLabel label;
  final String customLabel;
  final String formatted;
  final String building;
  final String floor;
  final String door;
  final String deliveryInstructions;
  final double? lat;
  final double? lng;
  final String? cityId;
  final bool isFavorite;
  final bool isDefault;
  final bool serviceable;
  final String recipientName;
  final String recipientPhone;
  final Map<String, dynamic> payload;

  static double? _asDouble(dynamic value) {
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value);
    return null;
  }

  factory Address.fromJson(Map<String, dynamic> json) {
    final label = addressLabelFromJson(json['label']?.toString());
    final customLabel = json['custom_label']?.toString() ?? '';
    final title = json['title']?.toString() ??
        json['name']?.toString() ??
        (label == AddressLabel.custom
            ? customLabel
            : addressLabelToJson(label));

    return Address(
      id: json['id']?.toString() ?? '',
      title: title,
      label: label,
      customLabel: customLabel,
      formatted: json['formatted']?.toString() ??
          json['line1']?.toString() ??
          json['address_line']?.toString() ??
          '',
      building: json['building']?.toString() ?? '',
      floor: json['floor']?.toString() ?? '',
      door: json['door']?.toString() ?? json['apartment']?.toString() ?? '',
      deliveryInstructions: json['delivery_instructions']?.toString() ??
          json['notes']?.toString() ??
          '',
      lat: _asDouble(json['lat'] ?? json['latitude']),
      lng: _asDouble(json['lng'] ?? json['lon'] ?? json['longitude']),
      cityId: json['city_id']?.toString(),
      isFavorite: json['is_favorite'] == true || json['isFavorite'] == true,
      isDefault: json['is_default'] == true || json['isDefault'] == true,
      serviceable: json['serviceable'] != false,
      recipientName: json['recipient_name']?.toString() ??
          json['recipientName']?.toString() ??
          '',
      recipientPhone: json['phone']?.toString() ??
          json['recipient_phone']?.toString() ??
          json['recipientPhone']?.toString() ??
          '',
      payload: Map<String, dynamic>.from(json),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'title': title,
        'label': addressLabelToJson(label),
        if (customLabel.isNotEmpty) 'custom_label': customLabel,
        'formatted': formatted,
        if (building.isNotEmpty) 'building': building,
        if (floor.isNotEmpty) 'floor': floor,
        if (door.isNotEmpty) 'door': door,
        if (deliveryInstructions.isNotEmpty)
          'delivery_instructions': deliveryInstructions,
        if (lat != null) 'lat': lat,
        if (lng != null) 'lng': lng,
        if (cityId != null) 'city_id': cityId,
        'is_favorite': isFavorite,
        'is_default': isDefault,
        'serviceable': serviceable,
        if (recipientName.isNotEmpty) 'recipient_name': recipientName,
        if (recipientPhone.isNotEmpty) 'phone': recipientPhone,
      };

  Address copyWith({
    String? id,
    String? title,
    AddressLabel? label,
    String? customLabel,
    String? formatted,
    String? building,
    String? floor,
    String? door,
    String? deliveryInstructions,
    double? lat,
    double? lng,
    String? cityId,
    bool? isFavorite,
    bool? isDefault,
    bool? serviceable,
    String? recipientName,
    String? recipientPhone,
    Map<String, dynamic>? payload,
  }) =>
      Address(
        id: id ?? this.id,
        title: title ?? this.title,
        label: label ?? this.label,
        customLabel: customLabel ?? this.customLabel,
        formatted: formatted ?? this.formatted,
        building: building ?? this.building,
        floor: floor ?? this.floor,
        door: door ?? this.door,
        deliveryInstructions:
            deliveryInstructions ?? this.deliveryInstructions,
        lat: lat ?? this.lat,
        lng: lng ?? this.lng,
        cityId: cityId ?? this.cityId,
        isFavorite: isFavorite ?? this.isFavorite,
        isDefault: isDefault ?? this.isDefault,
        serviceable: serviceable ?? this.serviceable,
        recipientName: recipientName ?? this.recipientName,
        recipientPhone: recipientPhone ?? this.recipientPhone,
        payload: payload ?? this.payload,
      );

  @override
  List<Object?> get props => [
        id,
        title,
        label,
        customLabel,
        formatted,
        building,
        floor,
        door,
        deliveryInstructions,
        lat,
        lng,
        cityId,
        isFavorite,
        isDefault,
        serviceable,
        recipientName,
        recipientPhone,
      ];
}
