import 'package:nexora_core/nexora_core.dart';

class AuthRemoteDataSource {
  const AuthRemoteDataSource(this._client);

  final ApiClient _client;

  Future<Result<void>> requestOtp(String phone) {
    return _client.post<void>(
      '/warehouse/auth/otp/request',
      data: {'phone': phone},
    );
  }

  Future<Result<void>> resendOtp(String phone) {
    return _client.post<void>(
      '/warehouse/auth/otp/resend',
      data: {'phone': phone},
    );
  }

  Future<Result<Map<String, dynamic>>> verifyOtp({
    required String phone,
    required String code,
  }) {
    return _client.post<Map<String, dynamic>>(
      '/warehouse/auth/otp/verify',
      data: {'phone': phone, 'code': code},
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<Map<String, dynamic>>> refresh(String refreshToken) {
    return _client.post<Map<String, dynamic>>(
      '/warehouse/auth/refresh',
      data: {'refresh_token': refreshToken},
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<Map<String, dynamic>>> getMe() {
    return _client.get<Map<String, dynamic>>(
      '/warehouse/me',
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<Map<String, dynamic>>> clockIn({
    String? stationId,
    String? storeId,
  }) {
    return _client.post<Map<String, dynamic>>(
      '/warehouse/shifts/clock-in',
      data: {
        if (stationId != null) 'station_id': stationId,
        if (storeId != null) 'store_id': storeId,
      },
      parser: (json) => Map<String, dynamic>.from(json as Map),
    );
  }

  Future<Result<void>> signOut() {
    return _client.post<void>('/warehouse/auth/logout');
  }
}
