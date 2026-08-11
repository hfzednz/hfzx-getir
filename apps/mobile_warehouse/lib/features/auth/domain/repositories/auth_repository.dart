import 'package:nexora_core/nexora_core.dart';

import '../entities/warehouse_session.dart';

abstract class AuthRepository {
  Future<Result<void>> requestOtp(String phone);
  Future<Result<void>> resendOtp(String phone);
  Future<Result<WarehouseSession>> verifyOtp({
    required String phone,
    required String code,
  });
  Future<Result<WarehouseSession>> refreshSession();
  Future<Result<WarehouseSession>> getMe();
  Future<Result<WarehouseSession>> clockIn({
    String? stationId,
    String? storeId,
  });
  Future<Result<bool>> authenticateWithBiometrics({
    String reason = 'Unlock warehouse actions',
  });
  Future<Result<void>> signOut();
}
