import '../../domain/entities/city_entity.dart';

class CityContextModel {
  const CityContextModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory CityContextModel.fromJson(Map<String, dynamic> json) => CityContextModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  CityContext toEntity() => CityContext(id: id, title: title, payload: raw);
}
