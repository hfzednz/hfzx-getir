import 'package:local_auth/local_auth.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/warehouse_session.dart';
import '../../domain/repositories/auth_repository.dart';
import '../datasources/auth_remote_datasource.dart';

class AuthRepositoryImpl implements AuthRepository {
  AuthRepositoryImpl(
    this._remote,
    this._tokenStore,
    this._prefs, {
    LocalAuthentication? localAuth,
  }) : _localAuth = localAuth ?? LocalAuthentication();

  final AuthRemoteDataSource _remote;
  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;
  final LocalAuthentication _localAuth;

  static const _userIdKey = 'warehouse_user_id';
  static const _displayNameKey = 'display_name';
  static const _phoneKey = 'phone';
  static const _roleKey = 'warehouse_role';
  static const _storeIdKey = 'store_id';
  static const _stationIdKey = 'station_id';
  static const _shiftIdKey = 'shift_id';
  static const _kycOkKey = 'kyc_ok';
  static const _deviceOkKey = 'device_ok';
  static const _sessionJsonKey = 'warehouse_session_json';

  @override
  Future<Result<void>> requestOtp(String phone) => _remote.requestOtp(phone);

  @override
  Future<Result<void>> resendOtp(String phone) => _remote.resendOtp(phone);

  @override
  Future<Result<WarehouseSession>> verifyOtp({
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
  Future<Result<WarehouseSession>> refreshSession() async {
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
  Future<Result<WarehouseSession>> getMe() async {
    final result = await _remote.getMe();
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    final access = await _tokenStore.readAccessToken() ?? '';
    final refresh = await _tokenStore.readRefreshToken() ?? '';
    return _persistSession({
      ...result.valueOrNull!,
      'access_token': access,
      'refresh_token': refresh,
    });
  }

  @override
  Future<Result<WarehouseSession>> clockIn({
    String? stationId,
    String? storeId,
  }) async {
    final result = await _remote.clockIn(
      stationId: stationId,
      storeId: storeId,
    );
    if (result.isFailure) {
      return Failure(result.errorOrNull!);
    }
    final access = await _tokenStore.readAccessToken() ?? '';
    final refresh = await _tokenStore.readRefreshToken() ?? '';
    final existing = _prefs.get<Map>(_sessionJsonKey);
    return _persistSession({
      if (existing != null) ...Map<String, dynamic>.from(existing),
      ...result.valueOrNull!,
      'access_token': access,
      'refresh_token': refresh,
      if (stationId != null) 'station_id': stationId,
      if (storeId != null) 'store_id': storeId,
    });
  }

  @override
  Future<Result<bool>> authenticateWithBiometrics({
    String reason = 'Unlock warehouse actions',
  }) async {
    try {
      final canCheck = await _localAuth.canCheckBiometrics ||
          await _localAuth.isDeviceSupported();
      if (!canCheck) {
        return const Failure(
          NexoraValidationException(
            code: NexoraErrorCode.validationFailed,
            message: 'Biometrics unavailable on this device',
          ),
        );
      }
      final ok = await _localAuth.authenticate(
        localizedReason: reason,
        options: const AuthenticationOptions(
          biometricOnly: true,
          stickyAuth: true,
        ),
      );
      return Success(ok);
    } catch (e) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: e.toString(),
        ),
      );
    }
  }

  @override
  Future<Result<void>> signOut() async {
    await _remote.signOut();
    await _tokenStore.clear();
    await _prefs.remove(_userIdKey);
    await _prefs.remove(_displayNameKey);
    await _prefs.remove(_phoneKey);
    await _prefs.remove(_roleKey);
    await _prefs.remove(_storeIdKey);
    await _prefs.remove(_stationIdKey);
    await _prefs.remove(_shiftIdKey);
    await _prefs.remove(_kycOkKey);
    await _prefs.remove(_deviceOkKey);
    await _prefs.remove(_sessionJsonKey);
    return const Success(null);
  }

  Future<Result<WarehouseSession>> _persistSession(
    Map<String, dynamic> json, {
    String? phone,
  }) async {
    final session = WarehouseSession.fromJson({
      ...json,
      if (phone != null) 'phone': phone,
    });
    if (session.accessToken.isNotEmpty && session.refreshToken.isNotEmpty) {
      await _tokenStore.saveTokens(
        accessToken: session.accessToken,
        refreshToken: session.refreshToken,
      );
    }
    await _prefs.set(_userIdKey, session.userId);
    await _prefs.set(_roleKey, session.role.wireName);
    await _prefs.set(_storeIdKey, session.storeId);
    await _prefs.set(_kycOkKey, session.kycOk);
    await _prefs.set(_deviceOkKey, session.deviceOk);
    if (session.displayName != null) {
      await _prefs.set(_displayNameKey, session.displayName!);
    }
    if (session.phone != null) {
      await _prefs.set(_phoneKey, session.phone!);
    }
    if (session.stationId != null) {
      await _prefs.set(_stationIdKey, session.stationId!);
    }
    if (session.shiftId != null) {
      await _prefs.set(_shiftIdKey, session.shiftId!);
    } else {
      await _prefs.remove(_shiftIdKey);
    }
    await _prefs.set(_sessionJsonKey, session.toJson());
    return Success(session);
  }
}
