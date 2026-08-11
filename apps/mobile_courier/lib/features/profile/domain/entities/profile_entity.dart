import 'package:equatable/equatable.dart';

import '../../../auth/domain/entities/courier_session.dart';

class CourierProfile extends Equatable {
  const CourierProfile({
    required this.courierId,
    this.displayName,
    this.phone,
    this.email,
    this.vehicleType,
    this.vehiclePlate,
    this.bankLast4,
    this.taxIdMasked,
    this.kycStatus = const KycStatus(),
  });

  final String courierId;
  final String? displayName;
  final String? phone;
  final String? email;
  final String? vehicleType;
  final String? vehiclePlate;
  final String? bankLast4;
  final String? taxIdMasked;
  final KycStatus kycStatus;

  factory CourierProfile.fromJson(Map<String, dynamic> json) {
    final kycRaw = json['kyc'] ?? json['kyc_status'] ?? json['documents'];
    return CourierProfile(
      courierId: json['courier_id']?.toString() ?? json['id']?.toString() ?? '',
      displayName: json['display_name']?.toString() ?? json['name']?.toString(),
      phone: json['phone']?.toString(),
      email: json['email']?.toString(),
      vehicleType: json['vehicle_type']?.toString() ??
          (json['vehicle'] is Map
              ? (json['vehicle'] as Map)['type']?.toString()
              : null),
      vehiclePlate: json['vehicle_plate']?.toString() ??
          (json['vehicle'] is Map
              ? (json['vehicle'] as Map)['plate']?.toString()
              : null),
      bankLast4: json['bank_last4']?.toString(),
      taxIdMasked: json['tax_id_masked']?.toString(),
      kycStatus: kycRaw is Map
          ? KycStatus.fromJson(Map<String, dynamic>.from(kycRaw))
          : const KycStatus(),
    );
  }

  @override
  List<Object?> get props => [
        courierId,
        displayName,
        phone,
        email,
        vehicleType,
        vehiclePlate,
        bankLast4,
        taxIdMasked,
        kycStatus,
      ];
}
