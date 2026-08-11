import 'package:equatable/equatable.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import 'auth_repository_provider.dart';
import '../../../../di/providers.dart';

enum AuthStatus { unknown, guest, authenticated, unauthenticated }

class AuthSession extends Equatable {
  const AuthSession({
    required this.status,
    this.userId,
    this.displayName,
    this.phone,
    this.email,
  });

  const AuthSession.unknown() : this(status: AuthStatus.unknown);
  const AuthSession.guest() : this(status: AuthStatus.guest);
  const AuthSession.unauthenticated()
      : this(status: AuthStatus.unauthenticated);

  final AuthStatus status;
  final String? userId;
  final String? displayName;
  final String? phone;
  final String? email;

  bool get isAuthenticated => status == AuthStatus.authenticated;
  bool get isGuest => status == AuthStatus.guest;
  bool get isGuestCheckoutAllowed => false;

  AuthSession copyWith({
    AuthStatus? status,
    String? userId,
    String? displayName,
    String? phone,
    String? email,
  }) {
    return AuthSession(
      status: status ?? this.status,
      userId: userId ?? this.userId,
      displayName: displayName ?? this.displayName,
      phone: phone ?? this.phone,
      email: email ?? this.email,
    );
  }

  @override
  List<Object?> get props => [status, userId, displayName, phone, email];
}

class AuthSessionNotifier extends StateNotifier<AuthSession> {
  AuthSessionNotifier(
    this._tokenStore,
    this._prefs, {
    Future<void> Function()? onLogout,
  })  : _onLogout = onLogout,
        super(const AuthSession.unknown());

  final SecureTokenStore _tokenStore;
  final PreferencesStore _prefs;
  final Future<void> Function()? _onLogout;

  static const _onboardingKey = 'onboarding_complete';
  static const _guestKey = 'guest_mode';

  Future<void> restore() async {
    final hasTokens = await _tokenStore.hasSession();
    if (hasTokens) {
      state = AuthSession(
        status: AuthStatus.authenticated,
        userId: _prefs.get<String>('user_id'),
        displayName: _prefs.get<String>('display_name'),
      );
      return;
    }
    if (_prefs.get<bool>(_guestKey) == true) {
      state = const AuthSession.guest();
      return;
    }
    state = const AuthSession.unauthenticated();
  }

  Future<void> setAuthenticated({
    required String userId,
    String? displayName,
    String? phone,
    String? email,
  }) async {
    await _prefs.set('user_id', userId);
    if (displayName != null) await _prefs.set('display_name', displayName);
    state = AuthSession(
      status: AuthStatus.authenticated,
      userId: userId,
      displayName: displayName,
      phone: phone,
      email: email,
    );
  }

  Future<void> continueAsGuest() async {
    await _prefs.set(_guestKey, true);
    state = const AuthSession.guest();
  }

  Future<void> signOut() async {
    await _onLogout?.call();
    await _tokenStore.clear();
    await _prefs.remove(_guestKey);
    await _prefs.remove('user_id');
    await _prefs.remove('display_name');
    await _prefs.remove(_biometricKey);
    state = const AuthSession.unauthenticated();
  }

  static const _biometricKey = 'biometric_enabled';

  bool get onboardingComplete => _prefs.get<bool>(_onboardingKey) ?? false;

  Future<void> completeOnboarding() => _prefs.set(_onboardingKey, true);
}

final authSessionProvider =
    StateNotifierProvider<AuthSessionNotifier, AuthSession>((ref) {
  return AuthSessionNotifier(
    ref.watch(tokenStoreProvider),
    ref.watch(preferencesStoreProvider),
    onLogout: () async {
      await ref.read(authRepositoryProvider).clearBiometric();
    },
  );
});

final onboardingCompleteProvider = Provider<bool>((ref) {
  return ref.watch(authSessionProvider.notifier).onboardingComplete;
});
