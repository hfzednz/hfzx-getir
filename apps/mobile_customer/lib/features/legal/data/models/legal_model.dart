import '../../domain/entities/legal_entity.dart';

class LegalDocumentModel {
  const LegalDocumentModel({required this.id, required this.title, required this.raw});

  final String id;
  final String title;
  final Map<String, dynamic> raw;

  factory LegalDocumentModel.fromJson(Map<String, dynamic> json) => LegalDocumentModel(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        raw: json,
      );

  LegalDocument toEntity() => LegalDocument(id: id, title: title, payload: raw);
}
