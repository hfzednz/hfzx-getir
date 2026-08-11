import 'package:nexora_core/nexora_core.dart';

import '../entities/device_session.dart';
import '../entities/privacy_controls.dart';
import '../repositories/auth_repository.dart';

class RegisterEmailUseCase {
  const RegisterEmailUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<AuthTokens>> call({
    required String email,
    required String password,
    String? name,
  }) =>
      _repository.registerEmail(email: email, password: password, name: name);
}

class ForgotPasswordUseCase {
  const ForgotPasswordUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call(String email) => _repository.forgotPassword(email);
}

class ResetPasswordUseCase {
  const ResetPasswordUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call({
    required String token,
    required String newPassword,
  }) =>
      _repository.resetPassword(token: token, newPassword: newPassword);
}

class VerifyEmailUseCase {
  const VerifyEmailUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call(String code) => _repository.verifyEmail(code);
}

class ResendOtpUseCase {
  const ResendOtpUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call(String phone) => _repository.resendOtp(phone);
}

class RefreshSessionUseCase {
  const RefreshSessionUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<AuthTokens>> call() => _repository.refreshSession();
}

class GuestSessionUseCase {
  const GuestSessionUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<AuthTokens>> call() => _repository.guestSession();
}

class ListDevicesUseCase {
  const ListDevicesUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<List<DeviceSession>>> call() => _repository.listDevices();
}

class RevokeDeviceUseCase {
  const RevokeDeviceUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call(String deviceId) =>
      _repository.revokeDevice(deviceId);
}

class DeleteAccountUseCase {
  const DeleteAccountUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call({String? reason}) =>
      _repository.deleteAccount(reason: reason);
}

class RequestDataExportUseCase {
  const RequestDataExportUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call() => _repository.requestDataExport();
}

class GetPrivacyControlsUseCase {
  const GetPrivacyControlsUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<PrivacyControls>> call() => _repository.getPrivacyControls();
}

class UpdatePrivacyControlsUseCase {
  const UpdatePrivacyControlsUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<PrivacyControls>> call(Map<String, dynamic> controls) =>
      _repository.updatePrivacyControls(controls);
}

class EnableBiometricUseCase {
  const EnableBiometricUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call() => _repository.enableBiometric();
}

class ClearBiometricUseCase {
  const ClearBiometricUseCase(this._repository);
  final AuthRepository _repository;

  Future<Result<void>> call() => _repository.clearBiometric();
}
