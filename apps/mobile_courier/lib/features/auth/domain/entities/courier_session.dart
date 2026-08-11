import 'package:equatable/equatable.dart';

enum KycDocumentType { license, vehicle, insurance }

enum KycDocumentStatus { missing, pending, approved, rejected }

class KycDocument extends Equatable {
  const KycDocument({
    required this.type,
    required this.status,
    this.rejectionReason,
    this.uploadedAt,
  });

  final KycDocumentType type;
  final KycDocumentStatus status;
  final String? rejectionReason;
  final DateTime? uploadedAt;

  bool get isApproved => status == KycDocumentStatus.approved;
  bool get needsUpload =>
      status == KycDocumentStatus.missing || status == KycDocumentStatus.rejected;

  factory KycDocument.fromJson(Map<String, dynamic> json) {
    return KycDocument(
      type: kycDocumentTypeFromJson(json['type']?.toString()),
      status: kycDocumentStatusFromJson(json['status']?.toString()),
      rejectionReason: json['rejection_reason']?.toString(),
      uploadedAt: json['uploaded_at'] != null
          ? DateTime.tryParse(json['uploaded_at'].toString())
          : null,
    );
  }

  Map<String, dynamic> toJson() => {
        'type': type.name,
        'status': status.name,
        if (rejectionReason != null) 'rejection_reason': rejectionReason,
        if (uploadedAt != null) 'uploaded_at': uploadedAt!.toIso8601String(),
      };

  @override
  List<Object?> get props => [type, status, rejectionReason, uploadedAt];
}

class KycStatus extends Equatable {
  const KycStatus({
    this.documents = const [],
    this.overallStatus = KycDocumentStatus.missing,
  });

  final List<KycDocument> documents;
  final KycDocumentStatus overallStatus;

  bool get isApproved =>
      overallStatus == KycDocumentStatus.approved ||
      (documents.isNotEmpty && documents.every((d) => d.isApproved));

  bool get isPending =>
      overallStatus == KycDocumentStatus.pending ||
      documents.any((d) => d.status == KycDocumentStatus.pending);

  KycDocument? document(KycDocumentType type) {
    for (final doc in documents) {
      if (doc.type == type) return doc;
    }
    return null;
  }

  factory KycStatus.fromJson(Map<String, dynamic> json) {
    final docsRaw = json['documents'] as List<dynamic>? ?? [];
    final docs = docsRaw
        .whereType<Map>()
        .map((e) => KycDocument.fromJson(Map<String, dynamic>.from(e)))
        .toList();
    return KycStatus(
      documents: docs,
      overallStatus: kycDocumentStatusFromJson(
        json['status']?.toString() ?? json['overall_status']?.toString(),
      ),
    );
  }

  Map<String, dynamic> toJson() => {
        'status': overallStatus.name,
        'documents': documents.map((d) => d.toJson()).toList(),
      };

  @override
  List<Object?> get props => [documents, overallStatus];
}

KycDocumentType kycDocumentTypeFromJson(String? raw) {
  switch (raw?.toLowerCase()) {
    case 'license':
    case 'driving_license':
      return KycDocumentType.license;
    case 'vehicle':
    case 'vehicle_registration':
      return KycDocumentType.vehicle;
    case 'insurance':
      return KycDocumentType.insurance;
    default:
      return KycDocumentType.license;
  }
}

KycDocumentStatus kycDocumentStatusFromJson(String? raw) {
  switch (raw?.toLowerCase()) {
    case 'pending':
    case 'submitted':
    case 'under_review':
      return KycDocumentStatus.pending;
    case 'approved':
    case 'verified':
      return KycDocumentStatus.approved;
    case 'rejected':
    case 'declined':
      return KycDocumentStatus.rejected;
    case 'missing':
    case 'not_uploaded':
    default:
      return KycDocumentStatus.missing;
  }
}

class CourierSession extends Equatable {
  const CourierSession({
    required this.courierId,
    required this.accessToken,
    required this.refreshToken,
    this.displayName,
    this.phone,
    this.cityId,
    this.kycStatus = const KycStatus(),
  });

  final String courierId;
  final String accessToken;
  final String refreshToken;
  final String? displayName;
  final String? phone;
  final String? cityId;
  final KycStatus kycStatus;

  factory CourierSession.fromJson(Map<String, dynamic> json) {
    final kycRaw = json['kyc'] ?? json['kyc_status'];
    return CourierSession(
      courierId: json['courier_id']?.toString() ??
          json['user_id']?.toString() ??
          json['id']?.toString() ??
          '',
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
      displayName: json['display_name']?.toString() ?? json['name']?.toString(),
      phone: json['phone']?.toString(),
      cityId: json['city_id']?.toString(),
      kycStatus: kycRaw is Map
          ? KycStatus.fromJson(Map<String, dynamic>.from(kycRaw))
          : const KycStatus(),
    );
  }

  Map<String, dynamic> toJson() => {
        'courier_id': courierId,
        'access_token': accessToken,
        'refresh_token': refreshToken,
        if (displayName != null) 'display_name': displayName,
        if (phone != null) 'phone': phone,
        if (cityId != null) 'city_id': cityId,
        'kyc': kycStatus.toJson(),
      };

  @override
  List<Object?> get props => [
        courierId,
        accessToken,
        refreshToken,
        displayName,
        phone,
        cityId,
        kycStatus,
      ];
}
