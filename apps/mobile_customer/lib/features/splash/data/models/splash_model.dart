import '../../domain/entities/splash_entity.dart';

class SplashStateModel {
  const SplashStateModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory SplashStateModel.fromJson(Map<String, dynamic> json) => SplashStateModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  SplashState toEntity() => SplashState(id: id, title: title, payload: raw);
}
