import 'package:equatable/equatable.dart';

class StoreSummary extends Equatable {
  const StoreSummary({
    required this.id,
    required this.name,
    this.status = 'open',
    this.etaMinutes,
    this.deliveryFeeMinor = 0,
    this.minOrderMinor = 0,
    this.category,
    this.imageUrl,
    this.open = true,
  });

  final String id;
  final String name;
  final String status;
  final int? etaMinutes;
  final int deliveryFeeMinor;
  final int minOrderMinor;
  final String? category;
  final String? imageUrl;
  final bool open;

  factory StoreSummary.fromJson(Map<String, dynamic> json) => StoreSummary(
        id: json['id']?.toString() ?? json['ID']?.toString() ?? '',
        name: json['name']?.toString() ??
            json['title']?.toString() ??
            json['Name']?.toString() ??
            '',
        status: json['status']?.toString() ?? 'open',
        etaMinutes: (json['etaMinutes'] as num?)?.toInt() ??
            (json['eta_minutes'] as num?)?.toInt(),
        deliveryFeeMinor: (json['deliveryFeeMinor'] as num?)?.toInt() ??
            (json['delivery_fee_minor'] as num?)?.toInt() ??
            0,
        minOrderMinor: (json['minOrderMinor'] as num?)?.toInt() ??
            (json['min_order_minor'] as num?)?.toInt() ??
            0,
        category: json['category']?.toString(),
        imageUrl: json['imageUrl']?.toString() ?? json['logo']?.toString(),
        open: json['open'] == true ||
            json['status']?.toString() == 'open' ||
            json['status']?.toString() == 'active',
      );

  @override
  List<Object?> get props => [id, name, status];
}
