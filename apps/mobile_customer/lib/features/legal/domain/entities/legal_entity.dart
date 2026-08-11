import 'package:equatable/equatable.dart';

class LegalDocument extends Equatable {
  const LegalDocument({
    required this.id,
    this.title = '',
    this.payload = const {},
  });

  final String id;
  final String title;
  final Map<String, dynamic> payload;

  factory LegalDocument.fromJson(Map<String, dynamic> json) => LegalDocument(
        id: json['id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        payload: Map<String, dynamic>.from(json),
      );

  Map<String, dynamic> toJson() => {'id': id, 'title': title, ...payload};

  @override
  List<Object?> get props => [id, title, payload];
}
