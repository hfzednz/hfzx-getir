import 'package:nexora_core/nexora_core.dart';

import '../entities/referral_entity.dart';

abstract class ReferralRepository {
  Future<Result<ReferralInfo>> fetchInfo();
  Future<Result<List<ReferralInvite>>> fetchInvites();
  Future<Result<ReferralInvite>> claimInvite({
    required String inviteCode,
    ReferralDeviceMetadata? device,
    String? idempotencyKey,
  });
}
