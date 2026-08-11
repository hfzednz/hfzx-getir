import 'package:nexora_core/nexora_core.dart';

import '../entities/courier_session.dart';

abstract class AuthRepository {
  Future<Result<void>> requestOtp(String phone);
  Future<Result<void>> resendOtp(String phone);
  Future<Result<CourierSession>> verifyOtp({
    required String phone,
    required String code,
  });
  Future<Result<CourierSession>> refreshSession();
  Future<Result<KycStatus>> getKycStatus();
  Future<Result<KycStatus>> uploadKycDocument({
    required KycDocumentType type,
    required String filePath,
  });
  Future<Result<void>> signOut();
}
