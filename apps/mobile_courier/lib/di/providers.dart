import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../data/local/courier_local_store.dart';
import '../shared/analytics/courier_analytics.dart';
import '../shared/location/location_service.dart';

final environmentProvider = Provider<AppEnvironment>((ref) {
  return AppEnvironment.fromEnv();
});

final tokenStoreProvider = Provider<SecureTokenStore>((ref) {
  return SecureTokenStore();
});

final preferencesStoreProvider = Provider<PreferencesStore>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final courierLocalStoreProvider = Provider<CourierLocalStore>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final mutationOutboxProvider = Provider<MutationOutbox>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final localeCodeProvider = StateProvider<String>((ref) => 'en');

final cityIdProvider = StateProvider<String?>((ref) => null);

final themeModeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.system);

final biometricEnabledProvider = StateProvider<bool>((ref) => false);

final tokenRefresherProvider = Provider<TokenRefresher>((ref) {
  final env = ref.watch(environmentProvider);
  final dio = Dio(BaseOptions(baseUrl: env.baseUrl));
  return _CourierTokenRefresher(dio);
});

final apiClientProvider = Provider<ApiClient>((ref) {
  final env = ref.watch(environmentProvider);
  final tokenStore = ref.watch(tokenStoreProvider);
  final refresher = ref.watch(tokenRefresherProvider);
  return ApiClient.create(
    config: ApiClientConfig(
      environment: env,
      languageProvider: () => ref.read(localeCodeProvider),
      cityIdProvider: () => ref.read(cityIdProvider),
    ),
    tokenStore: tokenStore,
    tokenRefresher: refresher,
  );
});

final connectivityServiceProvider = Provider<ConnectivityService>((ref) {
  return ConnectivityService();
});

final connectivityOnlineProvider = StreamProvider<bool>((ref) async* {
  final service = ref.watch(connectivityServiceProvider);
  yield await service.isOnline;
  await for (final _ in service.onConnectivityChanged) {
    yield await service.isOnline;
  }
});

final realtimeClientProvider = Provider<RealtimeClient>((ref) {
  final env = ref.watch(environmentProvider);
  final api = ref.watch(apiClientProvider);
  final client = RealtimeClient(
    wsBaseUrl: env.wsUrl,
    ticketProvider: () async {
      final result = await api.get<Map<String, dynamic>>(
        '/realtime/ticket',
        parser: (json) => Map<String, dynamic>.from(json as Map),
      );
      return result.fold(
        onSuccess: (data) =>
            data['ticket']?.toString() ??
            data['token']?.toString() ??
            '',
        onFailure: (error) => throw error,
      );
    },
  );
  ref.onDispose(() => client.dispose());
  return client;
});

final analyticsProvider = Provider<AnalyticsGateway>((ref) {
  return LoggingAnalyticsGateway();
});

final courierAnalyticsProvider = Provider<CourierAnalyticsTracker>((ref) {
  return CourierAnalyticsTracker(
    gateway: ref.watch(analyticsProvider),
    cityIdProvider: () => ref.read(cityIdProvider),
  );
});

final syncEngineProvider = Provider<SyncEngine>((ref) {
  return DefaultSyncEngine(
    outbox: ref.watch(mutationOutboxProvider),
    apiClient: ref.watch(apiClientProvider),
    connectivity: ref.watch(connectivityServiceProvider),
  );
});

final locationServiceProvider = Provider<LocationService>((ref) {
  final service = LocationService(
    localStore: ref.watch(courierLocalStoreProvider),
  );
  ref.onDispose(() => service.stop());
  return service;
});

final prefsProvider = preferencesStoreProvider;

class _CourierTokenRefresher implements TokenRefresher {
  _CourierTokenRefresher(this._dio);
  final Dio _dio;

  @override
  Future<TokenPair> refresh(String refreshToken) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/courier/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final data = response.data ?? {};
    return TokenPair(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }
}
