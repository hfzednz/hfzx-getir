import 'dart:async';
import 'dart:collection';

import 'package:dio/dio.dart';

import '../../config/nexora_headers.dart';
import '../../errors/dio_error_mapper.dart';
import '../../errors/error_envelope.dart';
import '../../errors/nexora_error_code.dart';
import '../../storage/secure_token_store.dart';

/// Token pair returned by refresh rotation (CONSTITUTION §29).
class TokenPair {
  const TokenPair({
    required this.accessToken,
    required this.refreshToken,
  });

  final String accessToken;
  final String refreshToken;
}

/// Performs refresh-token rotation against `POST /v1/auth/refresh`.
abstract class TokenRefresher {
  Future<TokenPair> refresh(String refreshToken);
}

/// Attaches Bearer auth, refreshes on 401 with request queue, revokes on reuse.
class AuthInterceptor extends QueuedInterceptor {
  AuthInterceptor({
    required SecureTokenStore tokenStore,
    required TokenRefresher tokenRefresher,
    required Dio refreshDio,
    this.refreshPath = '/auth/refresh',
    this.skipAuthPaths = const {'/auth/refresh', '/auth/login', '/auth/otp'},
  })  : _tokenStore = tokenStore,
        _tokenRefresher = tokenRefresher,
        _refreshDio = refreshDio;

  final SecureTokenStore _tokenStore;
  final TokenRefresher _tokenRefresher;
  final Dio _refreshDio;
  final String refreshPath;
  final Set<String> skipAuthPaths;

  Completer<void>? _refreshCompleter;
  final Queue<({RequestOptions options, ErrorInterceptorHandler handler})>
      _pending401 = Queue();

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    if (_shouldSkipAuth(options.path)) {
      handler.next(options);
      return;
    }

    final accessToken = await _tokenStore.readAccessToken();
    if (accessToken != null && accessToken.isNotEmpty) {
      options.headers[NexoraHeaders.authorization] = 'Bearer $accessToken';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode != 401 || _shouldSkipAuth(err.requestOptions.path)) {
      handler.next(err);
      return;
    }

    if (err.requestOptions.extra['auth_retry'] == true) {
      handler.next(err);
      return;
    }

    _pending401.add((options: err.requestOptions, handler: handler));

    if (_refreshCompleter != null) {
      await _refreshCompleter!.future;
      return _retryQueued();
    }

    _refreshCompleter = Completer<void>();

    try {
      await _rotateTokens();
      _refreshCompleter!.complete();
      await _retryQueued();
    } catch (error) {
      _refreshCompleter!.completeError(error);
      _failQueued(error);
    } finally {
      _refreshCompleter = null;
    }
  }

  Future<void> _rotateTokens() async {
    final refreshToken = await _tokenStore.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      await _tokenStore.clear();
      throw DioException(
        requestOptions: RequestOptions(path: refreshPath),
        message: 'No refresh token available',
      );
    }

    try {
      final pair = await _tokenRefresher.refresh(refreshToken);
      await _tokenStore.saveTokens(
        accessToken: pair.accessToken,
        refreshToken: pair.refreshToken,
      );
    } on DioException catch (error) {
      final mapped = DioErrorMapper.map(error);
      if (mapped.code == NexoraErrorCode.refreshTokenReused ||
          mapped.code == NexoraErrorCode.authInvalid) {
        await _tokenStore.clear();
      }
      rethrow;
    }
  }

  Future<void> _retryQueued() async {
    while (_pending401.isNotEmpty) {
      final pending = _pending401.removeFirst();
      final accessToken = await _tokenStore.readAccessToken();
      final options = pending.options;
      options.extra['auth_retry'] = true;
      if (accessToken != null && accessToken.isNotEmpty) {
        options.headers[NexoraHeaders.authorization] = 'Bearer $accessToken';
      }

      try {
        final response = await _refreshDio.fetch<dynamic>(options);
        pending.handler.resolve(response);
      } on DioException catch (error) {
        pending.handler.next(error);
      }
    }
  }

  void _failQueued(Object error) {
    while (_pending401.isNotEmpty) {
      final pending = _pending401.removeFirst();
      if (error is DioException) {
        pending.handler.next(error);
      } else {
        pending.handler.next(
          DioException(
            requestOptions: pending.options,
            error: error,
            message: error.toString(),
          ),
        );
      }
    }
  }

  bool _shouldSkipAuth(String path) {
    return skipAuthPaths.any((skip) => path.contains(skip));
  }
}

/// Default [TokenRefresher] using a dedicated Dio instance.
class DefaultTokenRefresher implements TokenRefresher {
  DefaultTokenRefresher({
    required Dio dio,
    this.refreshPath = '/auth/refresh',
  }) : _dio = dio;

  final Dio _dio;
  final String refreshPath;

  @override
  Future<TokenPair> refresh(String refreshToken) async {
    final response = await _dio.post<Map<String, dynamic>>(
      refreshPath,
      data: {'refresh_token': refreshToken},
      options: Options(extra: {'skip_auth': true}),
    );

    final data = response.data;
    if (data == null) {
      throw DioException(
        requestOptions: response.requestOptions,
        response: response,
        message: 'Empty refresh response',
      );
    }

    final access = data['access_token'] as String?;
    final refresh = data['refresh_token'] as String?;
    if (access == null || refresh == null) {
      throw DioException(
        requestOptions: response.requestOptions,
        response: response,
        message: 'Malformed refresh response',
      );
    }

    return TokenPair(accessToken: access, refreshToken: refresh);
  }
}

/// Parses refresh reuse from error envelope for explicit revoke flows.
bool isRefreshTokenReuse(DioException error) {
  final data = error.response?.data;
  if (data is Map<String, dynamic>) {
    final envelope = NexoraErrorEnvelope.fromJson(data);
    return envelope.code == NexoraErrorCode.refreshTokenReused;
  }
  return false;
}
