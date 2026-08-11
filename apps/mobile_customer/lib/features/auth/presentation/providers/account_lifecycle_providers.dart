import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/device_session.dart';
import '../../domain/entities/privacy_controls.dart';
import '../../domain/usecases/account_lifecycle_usecases.dart';
import '../providers/auth_repository_provider.dart';

final registerEmailUseCaseProvider = Provider(
  (ref) => RegisterEmailUseCase(ref.watch(authRepositoryProvider)),
);

final forgotPasswordUseCaseProvider = Provider(
  (ref) => ForgotPasswordUseCase(ref.watch(authRepositoryProvider)),
);

final resetPasswordUseCaseProvider = Provider(
  (ref) => ResetPasswordUseCase(ref.watch(authRepositoryProvider)),
);

final verifyEmailUseCaseProvider = Provider(
  (ref) => VerifyEmailUseCase(ref.watch(authRepositoryProvider)),
);

final resendOtpUseCaseProvider = Provider(
  (ref) => ResendOtpUseCase(ref.watch(authRepositoryProvider)),
);

final refreshSessionUseCaseProvider = Provider(
  (ref) => RefreshSessionUseCase(ref.watch(authRepositoryProvider)),
);

final guestSessionUseCaseProvider = Provider(
  (ref) => GuestSessionUseCase(ref.watch(authRepositoryProvider)),
);

final listDevicesUseCaseProvider = Provider(
  (ref) => ListDevicesUseCase(ref.watch(authRepositoryProvider)),
);

final revokeDeviceUseCaseProvider = Provider(
  (ref) => RevokeDeviceUseCase(ref.watch(authRepositoryProvider)),
);

final deleteAccountUseCaseProvider = Provider(
  (ref) => DeleteAccountUseCase(ref.watch(authRepositoryProvider)),
);

final requestDataExportUseCaseProvider = Provider(
  (ref) => RequestDataExportUseCase(ref.watch(authRepositoryProvider)),
);

final getPrivacyControlsUseCaseProvider = Provider(
  (ref) => GetPrivacyControlsUseCase(ref.watch(authRepositoryProvider)),
);

final updatePrivacyControlsUseCaseProvider = Provider(
  (ref) => UpdatePrivacyControlsUseCase(ref.watch(authRepositoryProvider)),
);

final enableBiometricUseCaseProvider = Provider(
  (ref) => EnableBiometricUseCase(ref.watch(authRepositoryProvider)),
);

final clearBiometricUseCaseProvider = Provider(
  (ref) => ClearBiometricUseCase(ref.watch(authRepositoryProvider)),
);

class AccountLifecycleState {
  const AccountLifecycleState({this.isLoading = false, this.error, this.message});

  final bool isLoading;
  final String? error;
  final String? message;

  AccountLifecycleState copyWith({
    bool? isLoading,
    String? error,
    String? message,
  }) {
    return AccountLifecycleState(
      isLoading: isLoading ?? this.isLoading,
      error: error,
      message: message,
    );
  }
}

class AccountLifecycleController extends StateNotifier<AccountLifecycleState> {
  AccountLifecycleController(this._ref) : super(const AccountLifecycleState());

  final Ref _ref;

  Future<bool> forgotPassword(String email) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(forgotPasswordUseCaseProvider).call(email);
    state = state.copyWith(
      isLoading: false,
      message: result.isSuccess ? 'Reset link sent' : null,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> resetPassword({
    required String token,
    required String newPassword,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(resetPasswordUseCaseProvider).call(
          token: token,
          newPassword: newPassword,
        );
    state = state.copyWith(
      isLoading: false,
      message: result.isSuccess ? 'Password updated' : null,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> deleteAccount({String? reason}) async {
    state = state.copyWith(isLoading: true, error: null);
    final result =
        await _ref.read(deleteAccountUseCaseProvider).call(reason: reason);
    state = state.copyWith(
      isLoading: false,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> requestDataExport() async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(requestDataExportUseCaseProvider).call();
    state = state.copyWith(
      isLoading: false,
      message: result.isSuccess ? 'Export requested' : null,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> revokeDevice(String deviceId) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(revokeDeviceUseCaseProvider).call(deviceId);
    state = state.copyWith(
      isLoading: false,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> updatePrivacy(PrivacyControls controls) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref
        .read(updatePrivacyControlsUseCaseProvider)
        .call(controls.toJson());
    state = state.copyWith(
      isLoading: false,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }

  Future<bool> resendOtp(String phone) async {
    state = state.copyWith(isLoading: true, error: null);
    final result = await _ref.read(resendOtpUseCaseProvider).call(phone);
    state = state.copyWith(
      isLoading: false,
      message: result.isSuccess ? 'Code resent' : null,
      error: result.isFailure ? result.errorOrNull?.message : null,
    );
    return result.isSuccess;
  }
}

final accountLifecycleControllerProvider =
    StateNotifierProvider<AccountLifecycleController, AccountLifecycleState>(
  (ref) => AccountLifecycleController(ref),
);

final devicesListProvider =
    FutureProvider.autoDispose<List<DeviceSession>>((ref) async {
  final result = await ref.watch(listDevicesUseCaseProvider).call();
  return result.fold(
    onSuccess: (devices) => devices,
    onFailure: (e) => throw e,
  );
});

final privacyControlsProvider =
    FutureProvider.autoDispose<PrivacyControls>((ref) async {
  final result = await ref.watch(getPrivacyControlsUseCaseProvider).call();
  return result.fold(
    onSuccess: (controls) => controls,
    onFailure: (e) => throw e,
  );
});
