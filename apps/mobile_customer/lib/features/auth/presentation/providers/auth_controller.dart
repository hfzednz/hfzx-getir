import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:local_auth/local_auth.dart';
import 'package:sign_in_with_apple/sign_in_with_apple.dart';

import '../../../../routing/route_names.dart';
import '../../../cart/presentation/providers/cart_providers.dart';
import '../../data/repositories/auth_repository_impl.dart';
import 'auth_repository_provider.dart';
import 'auth_session_provider.dart';

class AuthControllerState {
  const AuthControllerState({this.isLoading = false, this.error});

  final bool isLoading;
  final String? error;

  AuthControllerState copyWith({bool? isLoading, String? error}) {
    return AuthControllerState(
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class AuthController extends StateNotifier<AuthControllerState> {
  AuthController(this._ref) : super(const AuthControllerState());

  final Ref _ref;
  final _localAuth = LocalAuthentication();

  AuthRepositoryImpl get _repo => _ref.read(authRepositoryProvider);

  Future<void> _onAuthSuccess() => mergeCartAfterLogin(_ref);

  Future<bool> requestOtp(String phone) async {
    state = state.copyWith(isLoading: true);
    final result = await _repo.requestOtp(phone);
    state = state.copyWith(isLoading: false);
    return result.isSuccess;
  }

  Future<bool> verifyOtp({
    required String phone,
    required String code,
    required BuildContext context,
  }) async {
    state = state.copyWith(isLoading: true);
    final result = await _repo.verifyOtp(phone: phone, code: code);
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (session) async {
        await _ref.read(authSessionProvider.notifier).setAuthenticated(
              userId: session.userId,
              displayName: session.displayName,
              phone: phone,
            );
        await _onAuthSuccess();
        await _maybeEnableBiometric();
        return true;
      },
      onFailure: (_) => false,
    );
  }

  Future<bool> signInEmail({
    required String email,
    required String password,
    required BuildContext context,
  }) async {
    state = state.copyWith(isLoading: true);
    final result = await _repo.signInEmail(email: email, password: password);
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (session) async {
        await _ref.read(authSessionProvider.notifier).setAuthenticated(
              userId: session.userId,
              displayName: session.displayName,
              email: email,
            );
        await _onAuthSuccess();
        return true;
      },
      onFailure: (_) => false,
    );
  }

  Future<void> signInGoogle(BuildContext context) async {
    state = state.copyWith(isLoading: true);
    try {
      final account = await GoogleSignIn().signIn();
      if (account == null) {
        state = state.copyWith(isLoading: false);
        return;
      }
      final auth = await account.authentication;
      final result = await _repo.signInGoogle(idToken: auth.idToken ?? '');
      await result.fold(
        onSuccess: (session) async {
          await _ref.read(authSessionProvider.notifier).setAuthenticated(
                userId: session.userId,
                displayName: session.displayName ?? account.displayName,
                email: account.email,
              );
          await _onAuthSuccess();
          if (context.mounted) context.go(RouteNames.home);
        },
        onFailure: (_) {},
      );
    } finally {
      state = state.copyWith(isLoading: false);
    }
  }

  Future<void> signInApple(BuildContext context) async {
    state = state.copyWith(isLoading: true);
    try {
      final credential = await SignInWithApple.getAppleIDCredential(
        scopes: [
          AppleIDAuthorizationScopes.email,
          AppleIDAuthorizationScopes.fullName,
        ],
      );
      final result = await _repo.signInApple(
        identityToken: credential.identityToken ?? '',
      );
      await result.fold(
        onSuccess: (session) async {
          await _ref.read(authSessionProvider.notifier).setAuthenticated(
                userId: session.userId,
                displayName: session.displayName,
                email: credential.email,
              );
          await _onAuthSuccess();
          if (context.mounted) context.go(RouteNames.home);
        },
        onFailure: (_) {},
      );
    } finally {
      state = state.copyWith(isLoading: false);
    }
  }

  Future<void> continueAsGuest(BuildContext context) async {
    await _ref.read(authSessionProvider.notifier).continueAsGuest();
    if (context.mounted) context.go(RouteNames.home);
  }

  Future<void> signOut() async {
    await _repo.signOut();
    await _ref.read(authSessionProvider.notifier).signOut();
  }

  Future<void> _maybeEnableBiometric() async {
    final can = await _localAuth.canCheckBiometrics;
    if (can) {
      await _repo.enableBiometric();
    }
  }

  Future<bool> registerEmail({
    required String email,
    required String password,
    String? name,
    required BuildContext context,
  }) async {
    state = state.copyWith(isLoading: true);
    final result = await _repo.registerEmail(
      email: email,
      password: password,
      name: name,
    );
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (session) async {
        await _ref.read(authSessionProvider.notifier).setAuthenticated(
              userId: session.userId,
              displayName: session.displayName ?? name,
              email: email,
            );
        await _onAuthSuccess();
        return true;
      },
      onFailure: (_) => false,
    );
  }

  Future<bool> resendOtp(String phone) async {
    final result = await _repo.resendOtp(phone);
    return result.isSuccess;
  }
}

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthControllerState>((ref) {
  return AuthController(ref);
});
