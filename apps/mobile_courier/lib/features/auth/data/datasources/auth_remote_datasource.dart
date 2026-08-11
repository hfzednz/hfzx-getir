import 'package:dio/dio.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/courier_session.dart';

class AuthRemoteDataSource {
  const AuthRemoteDataSource(this._client);

  final ApiClient _client;

  Future<Result<void>> requestOtp(String phone) {
    return _client.post<void>(
      '/courier/auth/otp/request',
      data: {'phone': phone},
    );
  }

  Future<Result<void>> resendOtp(String phone) {
    return _client.post<void>(
      '/courier/auth/otp/resend',
      data: {'phone': phone},
    );
  }

  Future<Result<Map<String, dynamic>>> verifyOtp({
    required String phone,
    required String code,
  }) {
    return _client.post<Map<String, dynamic>>(
      '/courier/auth/otp/verify',
      data: {'phone': phone, 'code': code},
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<Map<String, dynamic>>> refresh(String refreshToken) {
    return _client.post<Map<String, dynamic>>(
      '/courier/auth/refresh',
      data: {'refresh_token': refreshToken},
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<KycStatus>> getKyc() {
    return _client.get<KycStatus>(
      '/courier/kyc',
      parser: (json) => KycStatus.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<KycStatus>> uploadKycDocument({
    required KycDocumentType type,
    required String filePath,
  }) async {
    final formData = FormData.fromMap({
      'type': type.name,
      'file': await MultipartFile.fromFile(filePath),
    });
    return _client.post<KycStatus>(
      '/courier/kyc',
      data: formData,
      parser: (json) => KycStatus.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<void>> signOut() {
    return _client.post<void>('/courier/auth/logout');
  }
}
