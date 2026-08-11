import '../../domain/entities/onboarding_entity.dart';

class OnboardingPageModel {
  const OnboardingPageModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory OnboardingPageModel.fromJson(Map<String, dynamic> json) => OnboardingPageModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  OnboardingPage toEntity() => OnboardingPage(id: id, title: title, payload: raw);
}
