import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/device_session.dart';
import '../../domain/entities/privacy_controls.dart';
import '../../domain/repositories/auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  const AuthRepositoryImpl(
    this._client,
    this._tokenStore,
    this._prefs,
  );

  final ApiClient _client;
  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;

  static const _biometricKey = 'biometric_enabled';

  @override
  Future<Result<void>> requestOtp(String phone) async {
    return _client.post<void>(
      '/auth/otp/request',
      data: {'phone': phone},
    );
  }

  @override
  Future<Result<void>> resendOtp(String phone) async {
    return _client.post<void>(
      '/auth/otp/resend',
      data: {'phone': phone},
    );
  }

  @override
  Future<Result<AuthTokens>> verifyOtp({
    required String phone,
    required String code,
  }) async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/otp/verify',
      data: {'phone': phone, 'code': code},
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<AuthTokens>> verifyPhone({
    required String phone,
    required String code,
  }) =>
      verifyOtp(phone: phone, code: code);

  @override
  Future<Result<AuthTokens>> signInEmail({
    required String email,
    required String password,
  }) async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/email/login',
      data: {'email': email, 'password': password},
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<AuthTokens>> registerEmail({
    required String email,
    required String password,
    String? name,
  }) async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/email/register',
      data: {
        'email': email,
        'password': password,
        if (name != null && name.isNotEmpty) 'name': name,
      },
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<void>> forgotPassword(String email) async {
    return _client.post<void>(
      '/auth/email/forgot-password',
      data: {'email': email},
    );
  }

  @override
  Future<Result<void>> resetPassword({
    required String token,
    required String newPassword,
  }) async {
    return _client.post<void>(
      '/auth/email/reset-password',
      data: {'token': token, 'new_password': newPassword},
    );
  }

  @override
  Future<Result<void>> verifyEmail(String code) async {
    return _client.post<void>(
      '/auth/email/verify',
      data: {'code': code},
    );
  }

  @override
  Future<Result<AuthTokens>> signInGoogle({required String idToken}) async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/social/google',
      data: {'id_token': idToken},
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<AuthTokens>> signInApple({required String identityToken}) async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/social/apple',
      data: {'identity_token': identityToken},
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<AuthTokens>> refreshSession() async {
    final refreshToken = await _tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      return const Failure(
        NexoraAuthException(
          code: NexoraErrorCode.authInvalid,
          message: 'No refresh token available',
        ),
      );
    }
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/refresh',
      data: {'refresh_token': refreshToken},
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<AuthTokens>> guestSession() async {
    final result = await _client.post<Map<String, dynamic>>(
      '/auth/guest',
      parser: (json) => json as Map<String, dynamic>,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistTokens(result.valueOrNull!);
  }

  @override
  Future<Result<List<DeviceSession>>> listDevices() async {
    return _client.get<List<DeviceSession>>(
      '/auth/devices',
      parser: (json) {
        final list = json as List<dynamic>? ?? [];
        return list
            .map((e) => DeviceSession.fromJson(e as Map<String, dynamic>))
            .toList();
      },
    );
  }

  @override
  Future<Result<void>> revokeDevice(String deviceId) async {
    return _client.delete<void>('/auth/devices/$deviceId');
  }

  @override
  Future<Result<void>> deleteAccount({String? reason}) async {
    return _client.post<void>(
      '/auth/account/delete',
      data: {if (reason != null && reason.isNotEmpty) 'reason': reason},
    );
  }

  @override
  Future<Result<void>> requestDataExport() async {
    return _client.post<void>('/auth/account/export');
  }

  @override
  Future<Result<PrivacyControls>> getPrivacyControls() async {
    return _client.get<PrivacyControls>(
      '/auth/privacy',
      parser: (json) =>
          PrivacyControls.fromJson(json as Map<String, dynamic>),
    );
  }

  @override
  Future<Result<PrivacyControls>> updatePrivacyControls(
    Map<String, dynamic> controls,
  ) async {
    return _client.patch<PrivacyControls>(
      '/auth/privacy',
      data: controls,
      parser: (json) =>
          PrivacyControls.fromJson(json as Map<String, dynamic>),
    );
  }

  @override
  Future<Result<void>> enableBiometric() async {
    await _prefs.set(_biometricKey, true);
    return const Success(null);
  }

  @override
  Future<Result<void>> clearBiometric() async {
    await _prefs.remove(_biometricKey);
    return const Success(null);
  }

  @override
  Future<Result<void>> signOut() async {
    await _client.post<void>('/auth/logout');
    await _tokenStore.clear();
    return const Success(null);
  }

  Future<Result<AuthTokens>> _persistTokens(Map<String, dynamic> json) async {
    final tokens = AuthTokens.fromJson(json);
    await _tokenStore.saveTokens(
      accessToken: tokens.accessToken,
      refreshToken: tokens.refreshToken,
    );
    return Success(tokens);
  }
}
