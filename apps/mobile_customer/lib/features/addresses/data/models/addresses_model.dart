import '../../domain/entities/addresses_entity.dart';

class AddressModel {
  const AddressModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory AddressModel.fromJson(Map<String, dynamic> json) => AddressModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ??
            json['name']?.toString() ??
            json['custom_label']?.toString() ??
            '',
        raw: json,
      );

  Address toEntity() => Address.fromJson(raw);
}
