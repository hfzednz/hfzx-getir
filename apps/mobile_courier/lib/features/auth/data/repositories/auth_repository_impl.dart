import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/courier_session.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';

class AuthRepositoryImpl implements AuthRepository {
  const AuthRepositoryImpl(
    this._remote,
    this._tokenStore,
    this._prefs,
  );

  final AuthRemoteDataSource _remote;
  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;

  static const _courierIdKey = 'courier_id';
  static const _displayNameKey = 'display_name';
  static const _phoneKey = 'phone';
  static const _kycKey = 'kyc_status_json';

  @override
  Future<Result<void>> requestOtp(String phone) => _remote.requestOtp(phone);

  @override
  Future<Result<void>> resendOtp(String phone) => _remote.resendOtp(phone);

  @override
  Future<Result<CourierSession>> verifyOtp({
    required String phone,
    required String code,
  }) async {
    final result = await _remote.verifyOtp(phone: phone, code: code);
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistSession(result.valueOrNull!, phone: phone);
  }

  @override
  Future<Result<CourierSession>> refreshSession() async {
    final refreshToken = await _tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      return const Failure(
        NexoraAuthException(
          code: NexoraErrorCode.authInvalid,
          message: 'No refresh token available',
        ),
      );
    }
    final result = await _remote.refresh(refreshToken);
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    return _persistSession(result.valueOrNull!);
  }

  @override
  Future<Result<KycStatus>> getKycStatus() async {
    final result = await _remote.getKyc();
    if (result.isSuccess) {
      await _prefs.set(_kycKey, result.valueOrNull!.toJson());
    }
    return result;
  }

  @override
  Future<Result<KycStatus>> uploadKycDocument({
    required KycDocumentType type,
    required String filePath,
  }) async {
    final result = await _remote.uploadKycDocument(
      type: type,
      filePath: filePath,
    );
    if (result.isSuccess) {
      await _prefs.set(_kycKey, result.valueOrNull!.toJson());
    }
    return result;
  }

  @override
  Future<Result<void>> signOut() async {
    await _remote.signOut();
    await _tokenStore.clear();
    await _prefs.remove(_courierIdKey);
    await _prefs.remove(_displayNameKey);
    await _prefs.remove(_phoneKey);
    await _prefs.remove(_kycKey);
    return const Success(null);
  }

  Future<Result<CourierSession>> _persistSession(
    Map<String, dynamic> json, {
    String? phone,
  }) async {
    final session = CourierSession.fromJson({
      ...json,
      if (phone != null) 'phone': phone,
    });
    await _tokenStore.saveTokens(
      accessToken: session.accessToken,
      refreshToken: session.refreshToken,
    );
    await _prefs.set(_courierIdKey, session.courierId);
    if (session.displayName != null) {
      await _prefs.set(_displayNameKey, session.displayName!);
    }
    if (session.phone != null) {
      await _prefs.set(_phoneKey, session.phone!);
    }
    await _prefs.set(_kycKey, session.kycStatus.toJson());
    return Success(session);
  }
}
