import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../data/local/warehouse_local_store.dart';
import '../shared/analytics/warehouse_analytics.dart';

export '../features/auth/presentation/providers/warehouse_session_provider.dart'
    show warehouseSessionProvider;

final environmentProvider = Provider<AppEnvironment>((ref) {
  return AppEnvironment.fromEnv();
});

final tokenStoreProvider = Provider<SecureTokenStore>((ref) {
  return SecureTokenStore();
});

final preferencesStoreProvider = Provider<PreferencesStore>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final warehouseLocalStoreProvider = Provider<WarehouseLocalStore>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final mutationOutboxProvider = Provider<MutationOutbox>((ref) {
  throw UnimplementedError('Override in bootstrap');
});

final localeCodeProvider = StateProvider<String>((ref) => 'en');

final storeIdProvider = StateProvider<String?>((ref) => null);

final themeModeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.system);

final biometricEnabledProvider = StateProvider<bool>((ref) => false);

final tokenRefresherProvider = Provider<TokenRefresher>((ref) {
  final env = ref.watch(environmentProvider);
  final dio = Dio(BaseOptions(baseUrl: env.baseUrl));
  return _WarehouseTokenRefresher(dio);
});

final apiClientProvider = Provider<ApiClient>((ref) {
  final env = ref.watch(environmentProvider);
  final tokenStore = ref.watch(tokenStoreProvider);
  final refresher = ref.watch(tokenRefresherProvider);
  return ApiClient.create(
    config: ApiClientConfig(
      environment: env,
      languageProvider: () => ref.read(localeCodeProvider),
      // Reuse city header slot for store-scoped BFF requests.
      cityIdProvider: () => ref.read(storeIdProvider),
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

final warehouseAnalyticsProvider = Provider<WarehouseAnalyticsTracker>((ref) {
  return WarehouseAnalyticsTracker(
    gateway: ref.watch(analyticsProvider),
    storeIdProvider: () => ref.read(storeIdProvider),
  );
});

final syncEngineProvider = Provider<SyncEngine>((ref) {
  return DefaultSyncEngine(
    outbox: ref.watch(mutationOutboxProvider),
    apiClient: ref.watch(apiClientProvider),
    connectivity: ref.watch(connectivityServiceProvider),
  );
});

final prefsProvider = preferencesStoreProvider;

class _WarehouseTokenRefresher implements TokenRefresher {
  _WarehouseTokenRefresher(this._dio);
  final Dio _dio;

  @override
  Future<TokenPair> refresh(String refreshToken) async {
    final response = await _dio.post<Map<String, dynamic>>(
      '/warehouse/auth/refresh',
      data: {'refresh_token': refreshToken},
    );
    final data = response.data ?? {};
    return TokenPair(
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }
}
