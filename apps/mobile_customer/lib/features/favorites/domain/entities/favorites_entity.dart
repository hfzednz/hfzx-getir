import 'package:equatable/equatable.dart';

enum FavoriteType { product, brand, category, search, store }

class FavoriteEntry extends Equatable {
  const FavoriteEntry({
    required this.id,
    required this.type,
    required this.targetId,
    this.title = '',
    this.subtitle = '',
    this.imageUrl,
    this.metadata = const {},
    this.addedAt,
    this.pendingSync = false,
  });

  final String id;
  final FavoriteType type;
  final String targetId;
  final String title;
  final String subtitle;
  final String? imageUrl;
  final Map<String, dynamic> metadata;
  final DateTime? addedAt;
  final bool pendingSync;

  factory FavoriteEntry.fromJson(Map<String, dynamic> json) => FavoriteEntry(
        id: json['id']?.toString() ?? '${json['type']}:${json['target_id']}',
        type: FavoriteType.values.asNameMap()[json['type']?.toString()] ?? FavoriteType.product,
        targetId: json['target_id']?.toString() ?? json['product_id']?.toString() ?? '',
        title: json['title']?.toString() ?? '',
        subtitle: json['subtitle']?.toString() ?? '',
        imageUrl: json['image_url']?.toString(),
        metadata: Map<String, dynamic>.from(json['metadata'] as Map? ?? {}),
        addedAt: json['added_at'] != null ? DateTime.tryParse(json['added_at'].toString()) : null,
        pendingSync: json['pending_sync'] as bool? ?? false,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type.name,
        'target_id': targetId,
        'title': title,
        if (subtitle.isNotEmpty) 'subtitle': subtitle,
        if (imageUrl != null) 'image_url': imageUrl,
        if (metadata.isNotEmpty) 'metadata': metadata,
        if (addedAt != null) 'added_at': addedAt!.toIso8601String(),
      };

  @override
  List<Object?> get props => [id, type, targetId, title, pendingSync];

  FavoriteEntry copyWith({bool? pendingSync}) => FavoriteEntry(
        id: id,
        type: type,
        targetId: targetId,
        title: title,
        subtitle: subtitle,
        imageUrl: imageUrl,
        metadata: metadata,
        addedAt: addedAt,
        pendingSync: pendingSync ?? this.pendingSync,
      );
}

/// Legacy product-only favorite item — kept for backward compatibility with old Drift table.
class FavoriteItem extends Equatable {
  const FavoriteItem({
    required this.id,
    this.title = '',
    this.imageUrl,
    this.unitPriceMinor,
    this.type = FavoriteType.product,
  });

  final String id;
  final String title;
  final String? imageUrl;
  final int? unitPriceMinor;
  final FavoriteType type;

  FavoriteEntry toEntry() => FavoriteEntry(
        id: id,
        type: type,
        targetId: id,
        title: title,
        imageUrl: imageUrl,
        metadata: {if (unitPriceMinor != null) 'unit_price_minor': unitPriceMinor},
      );

  factory FavoriteItem.fromJson(Map<String, dynamic> json) => FavoriteItem(
        id: json['id']?.toString() ?? json['product_id']?.toString() ?? '',
        title: json['title']?.toString() ?? json['name']?.toString() ?? '',
        imageUrl: json['image_url']?.toString(),
        unitPriceMinor: (json['unit_price_minor'] as num?)?.toInt(),
        type: FavoriteType.values.asNameMap()[json['type']?.toString()] ?? FavoriteType.product,
      );

  @override
  List<Object?> get props => [id, title, type];
}
