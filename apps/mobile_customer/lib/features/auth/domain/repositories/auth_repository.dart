import 'package:equatable/equatable.dart';
import 'package:nexora_core/nexora_core.dart';

import '../entities/device_session.dart';
import '../entities/privacy_controls.dart';

class AuthTokens extends Equatable {
  const AuthTokens({
    required this.userId,
    this.displayName,
    required this.accessToken,
    required this.refreshToken,
  });

  final String userId;
  final String? displayName;
  final String accessToken;
  final String refreshToken;

  factory AuthTokens.fromJson(Map<String, dynamic> json) {
    String pick(List<String> keys) {
      for (final key in keys) {
        final v = json[key];
        if (v == null) continue;
        final s = v.toString().trim();
        if (s.isNotEmpty && s != 'null') return s;
      }
      return '';
    }

    final name = pick(const ['displayName', 'display_name']);
    return AuthTokens(
      userId: pick(const [
        'customerId',
        'principalId',
        'user_id',
        'id',
        'CustomerID',
        'PrincipalID',
      ]),
      displayName: name.isEmpty ? null : name,
      accessToken: pick(const ['accessToken', 'access_token', 'AccessToken']),
      refreshToken: pick(const ['refreshToken', 'refresh_token', 'RefreshToken']),
    );
  }

  @override
  List<Object?> get props => [userId, displayName, accessToken, refreshToken];
}

abstract class AuthRepository {
  Future<Result<void>> requestOtp(String phone);
  Future<Result<void>> resendOtp(String phone);
  Future<Result<AuthTokens>> verifyOtp({
    required String phone,
    required String code,
  });
  Future<Result<AuthTokens>> verifyPhone({
    required String phone,
    required String code,
  });
  Future<Result<AuthTokens>> signInEmail({
    required String email,
    required String password,
  });
  Future<Result<AuthTokens>> registerEmail({
    required String email,
    required String password,
    String? name,
  });
  Future<Result<void>> forgotPassword(String email);
  Future<Result<void>> resetPassword({
    required String token,
    required String newPassword,
  });
  Future<Result<void>> verifyEmail(String code);
  Future<Result<AuthTokens>> signInGoogle({required String idToken});
  Future<Result<AuthTokens>> signInApple({required String identityToken});
  Future<Result<AuthTokens>> refreshSession();
  Future<Result<AuthTokens>> guestSession();
  Future<Result<List<DeviceSession>>> listDevices();
  Future<Result<void>> revokeDevice(String deviceId);
  Future<Result<void>> deleteAccount({String? reason});
  Future<Result<void>> requestDataExport();
  Future<Result<PrivacyControls>> getPrivacyControls();
  Future<Result<PrivacyControls>> updatePrivacyControls(
    Map<String, dynamic> controls,
  );
  Future<Result<void>> enableBiometric();
  Future<Result<void>> clearBiometric();
  Future<Result<void>> signOut();
}
