import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/referral_entity.dart';
import '../../domain/repositories/referral_repository.dart';
import '../datasources/referral_remote_datasource.dart';

class ReferralRepositoryImpl implements ReferralRepository {
  const ReferralRepositoryImpl(this._remote);
  final ReferralRemoteDataSource _remote;

  @override
  Future<Result<ReferralInfo>> fetchInfo() => _remote.fetchInfo();

  @override
  Future<Result<List<ReferralInvite>>> fetchInvites() => _remote.fetchInvites();

  @override
  Future<Result<ReferralInvite>> claimInvite({
    required String inviteCode,
    ReferralDeviceMetadata? device,
    String? idempotencyKey,
  }) =>
      _remote.claimInvite(
        inviteCode: inviteCode,
        device: device,
        idempotencyKey: idempotencyKey,
      );
}
