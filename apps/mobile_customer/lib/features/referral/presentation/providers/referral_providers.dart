import 'dart:io';

import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:package_info_plus/package_info_plus.dart';
import 'package:share_plus/share_plus.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/analytics_providers.dart';
import '../../../../di/providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../data/datasources/referral_remote_datasource.dart';
import '../../data/repositories/referral_repository_impl.dart';
import '../../domain/entities/referral_entity.dart';
import '../../domain/repositories/referral_repository.dart';

final referralRemoteDataSourceProvider = Provider<ReferralRemoteDataSource>((ref) {
  return ReferralRemoteDataSource(ref.watch(apiClientProvider));
});

final referralRepositoryProvider = Provider<ReferralRepository>((ref) {
  return ReferralRepositoryImpl(ref.watch(referralRemoteDataSourceProvider));
});

final referralInfoProvider = FutureProvider.autoDispose<ReferralInfo>((ref) async {
  final result = await ref.watch(referralRepositoryProvider).fetchInfo();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final referralInvitesProvider = FutureProvider.autoDispose<List<ReferralInvite>>((ref) async {
  final result = await ref.watch(referralRepositoryProvider).fetchInvites();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final referralClaimProvider =
    AsyncNotifierProvider<ReferralClaimController, ReferralInvite?>(
  ReferralClaimController.new,
);

class ReferralClaimController extends AsyncNotifier<ReferralInvite?> {
  static const _deviceIdKey = 'referral_device_id';

  @override
  Future<ReferralInvite?> build() async => null;

  Future<ReferralDeviceMetadata> _collectDeviceMetadata() async {
    final prefs = ref.read(prefsProvider);
    var deviceId = prefs.get<String>(_deviceIdKey);
    if (deviceId == null || deviceId.isEmpty) {
      deviceId = const Uuid().v4();
      await prefs.set(_deviceIdKey, deviceId);
    }

    String? model;
    String? os;
    try {
      final plugin = DeviceInfoPlugin();
      if (Platform.isAndroid) {
        final info = await plugin.androidInfo;
        model = info.model;
        os = 'Android ${info.version.release}';
      } else if (Platform.isIOS) {
        final info = await plugin.iosInfo;
        model = info.utsname.machine;
        os = 'iOS ${info.systemVersion}';
      } else {
        os = Platform.operatingSystem;
        model = Platform.operatingSystemVersion;
      }
    } catch (_) {
      os = Platform.operatingSystem;
    }

    String? appVersion;
    try {
      final package = await PackageInfo.fromPlatform();
      appVersion = '${package.version}+${package.buildNumber}';
    } catch (_) {}

    return ReferralDeviceMetadata(
      deviceId: deviceId,
      model: model,
      os: os,
      appVersion: appVersion,
    );
  }

  Future<void> claim(String inviteCode) async {
    final code = inviteCode.trim().toUpperCase();
    if (code.isEmpty) {
      state = AsyncError(ArgumentError('Invite code is required'), StackTrace.current);
      return;
    }

    state = const AsyncLoading();
    state = await AsyncValue.guard(() async {
      final device = await _collectDeviceMetadata();
      final result = await ref.read(referralRepositoryProvider).claimInvite(
            inviteCode: code,
            device: device,
            idempotencyKey: const Uuid().v4(),
          );
      return result.fold(
        onSuccess: (invite) {
          ref.invalidate(referralInvitesProvider);
          ref.invalidate(referralInfoProvider);
          return invite;
        },
        onFailure: (e) => throw e,
      );
    });
  }
}

Future<void> shareReferralInvite(WidgetRef ref, ReferralInfo info) async {
  final text = info.shareUrl.isNotEmpty ? '${info.shareMessage}\n${info.shareUrl}' : info.shareMessage;
  await Share.share(text, subject: 'Join NEXORA');
  await ref.read(analyticsTrackerProvider).trackRaw(
        eventName: AnalyticsEvents.referralShared,
        props: {'invite_code': info.inviteCode},
      );
}
