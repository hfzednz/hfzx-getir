import 'package:equatable/equatable.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/providers.dart';
import '../../domain/entities/courier_session.dart';

enum AuthStatus { unknown, authenticated, unauthenticated }

class AuthSessionState extends Equatable {
  const AuthSessionState({
    required this.status,
    this.courierId,
    this.displayName,
    this.phone,
    this.cityId,
    this.kycStatus = const KycStatus(),
  });

  const AuthSessionState.unknown() : this(status: AuthStatus.unknown);

  const AuthSessionState.unauthenticated()
      : this(status: AuthStatus.unauthenticated);

  final AuthStatus status;
  final String? courierId;
  final String? displayName;
  final String? phone;
  final String? cityId;
  final KycStatus kycStatus;

  bool get isAuthenticated => status == AuthStatus.authenticated;

  AuthSessionState copyWith({
    AuthStatus? status,
    String? courierId,
    String? displayName,
    String? phone,
    String? cityId,
    KycStatus? kycStatus,
  }) {
    return AuthSessionState(
      status: status ?? this.status,
      courierId: courierId ?? this.courierId,
      displayName: displayName ?? this.displayName,
      phone: phone ?? this.phone,
      cityId: cityId ?? this.cityId,
      kycStatus: kycStatus ?? this.kycStatus,
    );
  }

  @override
  List<Object?> get props =>
      [status, courierId, displayName, phone, cityId, kycStatus];
}

class AuthSessionNotifier extends StateNotifier<AuthSessionState> {
  AuthSessionNotifier(
    this._tokenStore,
    this._prefs,
  ) : super(const AuthSessionState.unknown());

  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;

  static const _courierIdKey = 'courier_id';
  static const _displayNameKey = 'display_name';
  static const _phoneKey = 'phone';
  static const _kycKey = 'kyc_status_json';

  Future<void> restore() async {
    final hasTokens = await _tokenStore.hasSession();
    if (!hasTokens) {
      state = const AuthSessionState.unauthenticated();
      return;
    }

    final kycRaw = _prefs.get<Map>(_kycKey);
    final kyc = kycRaw != null
        ? KycStatus.fromJson(Map<String, dynamic>.from(kycRaw))
        : const KycStatus();

    state = AuthSessionState(
      status: AuthStatus.authenticated,
      courierId: _prefs.get<String>(_courierIdKey),
      displayName: _prefs.get<String>(_displayNameKey),
      phone: _prefs.get<String>(_phoneKey),
      kycStatus: kyc,
    );
  }

  Future<void> setAuthenticated(CourierSession session) async {
    if (session.cityId != null) {
      // city wired via caller reading cityIdProvider
    }
    state = AuthSessionState(
      status: AuthStatus.authenticated,
      courierId: session.courierId,
      displayName: session.displayName,
      phone: session.phone,
      cityId: session.cityId,
      kycStatus: session.kycStatus,
    );
  }

  Future<void> updateKyc(KycStatus kyc) async {
    await _prefs.set(_kycKey, kyc.toJson());
    state = state.copyWith(kycStatus: kyc);
  }

  Future<void> signOut() async {
    await _tokenStore.clear();
    await _prefs.remove(_courierIdKey);
    await _prefs.remove(_displayNameKey);
    await _prefs.remove(_phoneKey);
    await _prefs.remove(_kycKey);
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
