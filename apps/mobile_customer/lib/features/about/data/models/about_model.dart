import '../../domain/entities/about_entity.dart';

class AboutInfoModel {
  const AboutInfoModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory AboutInfoModel.fromJson(Map<String, dynamic> json) => AboutInfoModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  AboutInfo toEntity() => AboutInfo(id: id, title: title, payload: raw);
}
