import 'package:equatable/equatable.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/providers.dart';
import '../../domain/entities/warehouse_session.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthSessionState extends Equatable {
  const AuthSessionState({
    required this.status,
    this.userId,
    this.displayName,
    this.phone,
    this.role = WarehouseRole.picker,
    this.storeId,
    this.stationId,
    this.shiftId,
    this.kycOk = false,
    this.deviceOk = false,
  });

  const AuthSessionState.unknown() : this(status: AuthStatus.unknown);

  const AuthSessionState.unauthenticated()
      : this(status: AuthStatus.unauthenticated);

  final AuthStatus status;
  final String? userId;
  final String? displayName;
  final String? phone;
  final WarehouseRole role;
  final String? storeId;
  final String? stationId;
  final String? shiftId;
  final bool kycOk;
  final bool deviceOk;

  bool get isAuthenticated => status == AuthStatus.authenticated;

  bool get hasActiveShift => shiftId != null && shiftId!.isNotEmpty;

  bool get canEnterShell => isAuthenticated && kycOk && deviceOk && hasActiveShift;

  AuthSessionState copyWith({
    AuthStatus? status,
    String? userId,
    String? displayName,
    String? phone,
    WarehouseRole? role,
    String? storeId,
    String? stationId,
    String? shiftId,
    bool? kycOk,
    bool? deviceOk,
    bool clearShiftId = false,
  }) {
    return AuthSessionState(
      status: status ?? this.status,
      userId: userId ?? this.userId,
      displayName: displayName ?? this.displayName,
      phone: phone ?? this.phone,
      role: role ?? this.role,
      storeId: storeId ?? this.storeId,
      stationId: stationId ?? this.stationId,
      shiftId: clearShiftId ? null : (shiftId ?? this.shiftId),
      kycOk: kycOk ?? this.kycOk,
      deviceOk: deviceOk ?? this.deviceOk,
    );
  }

  @override
  List<Object?> get props => [
        status,
        userId,
        displayName,
        phone,
        role,
        storeId,
        stationId,
        shiftId,
        kycOk,
        deviceOk,
      ];
}

class AuthSessionNotifier extends StateNotifier<AuthSessionState> {
  AuthSessionNotifier(
    this._tokenStore,
    this._prefs,
  ) : super(const AuthSessionState.unknown());

  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;

  static const _userIdKey = 'warehouse_user_id';
  static const _displayNameKey = 'display_name';
  static const _phoneKey = 'phone';
  static const _roleKey = 'warehouse_role';
  static const _storeIdKey = 'store_id';
  static const _stationIdKey = 'station_id';
  static const _shiftIdKey = 'shift_id';
  static const _kycOkKey = 'kyc_ok';
  static const _deviceOkKey = 'device_ok';

  Future<void> restore() async {
    final hasTokens = await _tokenStore.hasSession();
    if (!hasTokens) {
      state = const AuthSessionState.unauthenticated();
      return;
    }

    state = AuthSessionState(
      status: AuthStatus.authenticated,
      userId: _prefs.get<String>(_userIdKey),
      displayName: _prefs.get<String>(_displayNameKey),
      phone: _prefs.get<String>(_phoneKey),
      role: WarehouseRole.fromString(_prefs.get<String>(_roleKey)),
      storeId: _prefs.get<String>(_storeIdKey),
      stationId: _prefs.get<String>(_stationIdKey),
      shiftId: _prefs.get<String>(_shiftIdKey),
      kycOk: _prefs.get<bool>(_kycOkKey) ?? false,
      deviceOk: _prefs.get<bool>(_deviceOkKey) ?? false,
    );
  }

  Future<void> setAuthenticated(WarehouseSession session) async {
    state = AuthSessionState(
      status: AuthStatus.authenticated,
      userId: session.userId,
      displayName: session.displayName,
      phone: session.phone,
      role: session.role,
      storeId: session.storeId,
      stationId: session.stationId,
      shiftId: session.shiftId,
      kycOk: session.kycOk,
      deviceOk: session.deviceOk,
    );
  }

  Future<void> updateShift({
    required String shiftId,
    String? stationId,
  }) async {
    await _prefs.set(_shiftIdKey, shiftId);
    if (stationId != null) {
      await _prefs.set(_stationIdKey, stationId);
    }
    state = state.copyWith(shiftId: shiftId, stationId: stationId);
  }

  Future<void> clearShift() async {
    await _prefs.remove(_shiftIdKey);
    state = state.copyWith(clearShiftId: true);
  }

  Future<void> signOut() async {
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
    state = const AuthSessionState.unauthenticated();
  }
}

final authSessionProvider =
    StateNotifierProvider<AuthSessionNotifier, AuthSessionState>((ref) {
  return AuthSessionNotifier(
    ref.watch(tokenStoreProvider),
    ref.watch(preferencesStoreProvider),
  );
});
