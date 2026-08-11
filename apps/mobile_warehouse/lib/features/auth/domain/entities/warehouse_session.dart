import 'package:equatable/equatable.dart';

import '../../../../shared/business_rules/warehouse_role.dart';

export '../../../../shared/business_rules/warehouse_role.dart';

class WarehouseSession extends Equatable {
  const WarehouseSession({
    required this.userId,
    required this.accessToken,
    required this.refreshToken,
    required this.role,
    required this.storeId,
    this.stationId,
    this.shiftId,
    this.displayName,
    this.phone,
    this.kycOk = false,
    this.deviceOk = false,
  });

  final String userId;
  final String accessToken;
  final String refreshToken;
  final WarehouseRole role;
  final String storeId;
  final String? stationId;
  final String? shiftId;
  final String? displayName;
  final String? phone;
  final bool kycOk;
  final bool deviceOk;

  bool get hasActiveShift => shiftId != null && shiftId!.isNotEmpty;

  bool get canEnterShell => kycOk && deviceOk && hasActiveShift;

  factory WarehouseSession.fromJson(Map<String, dynamic> json) {
    return WarehouseSession(
      userId: json['user_id']?.toString() ??
          json['warehouse_user_id']?.toString() ??
          json['id']?.toString() ??
          '',
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
      role: WarehouseRole.fromString(
        json['role']?.toString() ?? json['warehouse_role']?.toString(),
      ),
      storeId: json['store_id']?.toString() ?? '',
      stationId: json['station_id']?.toString(),
      shiftId: json['shift_id']?.toString() ??
          json['active_shift_id']?.toString(),
      displayName:
          json['display_name']?.toString() ?? json['name']?.toString(),
      phone: json['phone']?.toString(),
      kycOk: json['kyc_ok'] == true ||
          json['kyc_status']?.toString().toLowerCase() == 'approved',
      deviceOk: json['device_ok'] == true ||
          json['device_authorized'] == true,
    );
  }

  Map<String, dynamic> toJson() => {
        'user_id': userId,
        'access_token': accessToken,
        'refresh_token': refreshToken,
        'role': role.wireName,
        'store_id': storeId,
        if (stationId != null) 'station_id': stationId,
        if (shiftId != null) 'shift_id': shiftId,
        if (displayName != null) 'display_name': displayName,
        if (phone != null) 'phone': phone,
        'kyc_ok': kycOk,
        'device_ok': deviceOk,
      };

  WarehouseSession copyWith({
    String? userId,
    String? accessToken,
    String? refreshToken,
    WarehouseRole? role,
    String? storeId,
    String? stationId,
    String? shiftId,
    String? displayName,
    String? phone,
    bool? kycOk,
    bool? deviceOk,
    bool clearShiftId = false,
  }) {
    return WarehouseSession(
      userId: userId ?? this.userId,
      accessToken: accessToken ?? this.accessToken,
      refreshToken: refreshToken ?? this.refreshToken,
      role: role ?? this.role,
      storeId: storeId ?? this.storeId,
      stationId: stationId ?? this.stationId,
      shiftId: clearShiftId ? null : (shiftId ?? this.shiftId),
      displayName: displayName ?? this.displayName,
      phone: phone ?? this.phone,
      kycOk: kycOk ?? this.kycOk,
      deviceOk: deviceOk ?? this.deviceOk,
    );
  }

  @override
  List<Object?> get props => [
        userId,
        accessToken,
        refreshToken,
        role,
        storeId,
        stationId,
        shiftId,
        displayName,
        phone,
        kycOk,
        deviceOk,
      ];
}
