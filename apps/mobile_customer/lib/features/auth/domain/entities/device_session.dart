import 'package:equatable/equatable.dart';

class DeviceSession extends Equatable {
  const DeviceSession({
    required this.id,
    required this.label,
    this.platform,
    this.lastActiveAt,
    this.isCurrent = false,
  });

  final String id;
  final String label;
  final String? platform;
  final DateTime? lastActiveAt;
  final bool isCurrent;

  factory DeviceSession.fromJson(Map<String, dynamic> json) => DeviceSession(
        id: json['id']?.toString() ?? json['device_id']?.toString() ?? '',
        label: json['label']?.toString() ??
            json['name']?.toString() ??
            'Unknown device',
        platform: json['platform']?.toString(),
        lastActiveAt: json['last_active_at'] != null
            ? DateTime.tryParse(json['last_active_at'].toString())
            : null,
        isCurrent: json['is_current'] == true,
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'label': label,
        if (platform != null) 'platform': platform,
        if (lastActiveAt != null)
          'last_active_at': lastActiveAt!.toUtc().toIso8601String(),
        'is_current': isCurrent,
      };

  @override
  List<Object?> get props => [id, label, platform, lastActiveAt, isCurrent];
}
