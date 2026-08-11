import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../domain/entities/warehouse_session.dart';
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
    if (result.isFailure) {
      state = state.copyWith(error: result.errorOrNull?.message);
      return false;
    }
    return true;
  }

  Future<bool> resendOtp(String phone) async {
    final result = await _ref.read(authRepositoryProvider).resendOtp(phone);
    return result.isSuccess;
  }

  Future<WarehouseSession?> verifyOtp({
    required String phone,
    required String code,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).verifyOtp(
          phone: phone,
          code: code,
        );
    state = state.copyWith(isLoading: false);
    if (result.isFailure) {
      state = state.copyWith(error: result.errorOrNull?.message);
      return null;
    }
    final session = result.valueOrNull!;
    await _ref.read(authSessionProvider.notifier).setAuthenticated(session);
    if (session.storeId.isNotEmpty) {
      _ref.read(storeIdProvider.notifier).state = session.storeId;
    }
    return session;
  }

  Future<WarehouseSession?> refreshMe() async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(authRepositoryProvider).getMe();
    state = state.copyWith(isLoading: false);
    if (result.isFailure) {
      state = state.copyWith(error: result.errorOrNull?.message);
      return null;
    }
    final session = result.valueOrNull!;
    await _ref.read(authSessionProvider.notifier).setAuthenticated(session);
    if (session.storeId.isNotEmpty) {
      _ref.read(storeIdProvider.notifier).state = session.storeId;
    }
    return session;
  }

  Future<WarehouseSession?> clockIn({String? stationId}) async {
    state = state.copyWith(isLoading: true, error: null);
    final storeId = _ref.read(storeIdProvider);
    final result = await _ref.read(authRepositoryProvider).clockIn(
          stationId: stationId,
          storeId: storeId,
        );
    state = state.copyWith(isLoading: false);
    if (result.isFailure) {
      state = state.copyWith(error: result.errorOrNull?.message);
      return null;
    }
    final session = result.valueOrNull!;
    await _ref.read(authSessionProvider.notifier).setAuthenticated(session);
    if (session.storeId.isNotEmpty) {
      _ref.read(storeIdProvider.notifier).state = session.storeId;
    }
    return session;
  }

  Future<bool> unlockWithBiometrics({String? reason}) async {
    final result = await _ref.read(authRepositoryProvider).authenticateWithBiometrics(
          reason: reason ?? 'Unlock warehouse actions',
        );
    return result.valueOrNull ?? false;
  }

  Future<void> signOut() async {
    await _ref.read(authRepositoryProvider).signOut();
    await _ref.read(authSessionProvider.notifier).signOut();
    _ref.read(storeIdProvider.notifier).state = null;
  }
}

final authControllerProvider =
    StateNotifierProvider<AuthController, AuthControllerState>((ref) {
  return AuthController(ref);
});
