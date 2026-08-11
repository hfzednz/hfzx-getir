import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../domain/entities/courier_session.dart';
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

  Future<bool> requestOtp(String phone) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).requestOtp(phone);
    state = state.copyWith(isLoading: false);
    return result.isSuccess;
  }

  Future<bool> resendOtp(String phone) async {
    final result = await _ref.read(authRepositoryProvider).resendOtp(phone);
    return result.isSuccess;
  }

  Future<CourierSession?> verifyOtp({
    required String phone,
    required String code,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).verifyOtp(
          phone: phone,
          code: code,
        );
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (session) async {
        await _ref.read(authSessionProvider.notifier).setAuthenticated(session);
        if (session.cityId != null) {
          _ref.read(cityIdProvider.notifier).state = session.cityId;
        }
        return session;
      },
      onFailure: (error) {
        state = state.copyWith(error: error.message);
        return null;
      },
    );
  }

  Future<KycStatus?> refreshKyc() async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).getKycStatus();
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (kyc) async {
        await _ref.read(authSessionProvider.notifier).updateKyc(kyc);
        return kyc;
      },
      onFailure: (error) {
        state = state.copyWith(error: error.message);
        return null;
      },
    );
  }

  Future<KycStatus?> uploadKycDocument({
    required KycDocumentType type,
    required String filePath,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).uploadKycDocument(
          type: type,
          filePath: filePath,
        );
    state = state.copyWith(isLoading: false);
    return result.fold(
      onSuccess: (kyc) async {
        await _ref.read(authSessionProvider.notifier).updateKyc(kyc);
        return kyc;
      },
      onFailure: (error) {
        state = state.copyWith(error: error.message);
        return null;
      },
    );
  }

  Future<void> signOut() async {
    await _ref.read(authRepositoryProvider).signOut();
    await _ref.read(authSessionProvider.notifier).signOut();
  }
}

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthControllerState>((ref) {
  return AuthController(ref);
});
