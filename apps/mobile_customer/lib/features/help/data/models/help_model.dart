import '../../domain/entities/help_entity.dart';

class HelpArticleModel {
  const HelpArticleModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory HelpArticleModel.fromJson(Map<String, dynamic> json) => HelpArticleModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  HelpArticle toEntity() => HelpArticle(id: id, title: title, payload: raw);
}
