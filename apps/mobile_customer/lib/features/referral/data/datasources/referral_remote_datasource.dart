import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/referral_entity.dart';
import '../models/referral_model.dart';

class ReferralRemoteDataSource {
  const ReferralRemoteDataSource(this._client);
  final ApiClient _client;

  static const _infoPath = '/referral';
  static const _invitesPath = '/referral/invites';
  static const _claimPath = '/referral/claim';

  Future<Result<ReferralInfo>> fetchInfo() async {
    return _client.get<ReferralInfo>(
      _infoPath,
      parser: (json) => ReferralInfoModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<ReferralInvite>>> fetchInvites() async {
    return _client.get<List<ReferralInvite>>(
      _invitesPath,
      parser: (json) => (json as List<dynamic>)
          .map((e) => ReferralInviteModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<ReferralInvite>> claimInvite({
    required String inviteCode,
    ReferralDeviceMetadata? device,
    String? idempotencyKey,
  }) async {
    return _client.post<ReferralInvite>(
      _claimPath,
      data: {
        'invite_code': inviteCode,
        if (device != null) 'device': device.toJson(),
      },
      idempotencyKey: idempotencyKey,
      parser: (json) => ReferralInviteModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
