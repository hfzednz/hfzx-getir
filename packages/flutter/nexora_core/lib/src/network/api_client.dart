import 'package:dio/dio.dart';
import 'package:logging/logging.dart';

import '../config/app_environment.dart';
import '../errors/dio_error_mapper.dart';
import '../errors/nexora_exception.dart';
import '../errors/result.dart';
import 'certificate_pinning_adapter.dart';
import 'interceptors/accept_language_interceptor.dart';
import 'interceptors/auth_interceptor.dart';
import 'interceptors/city_header_interceptor.dart';
import 'interceptors/idempotency_interceptor.dart';
import 'interceptors/logging_interceptor.dart';
import '../storage/secure_token_store.dart';

export 'interceptors/auth_interceptor.dart'
    show AuthInterceptor, DefaultTokenRefresher, TokenPair, TokenRefresher;

/// Configuration for [ApiClient].
class ApiClientConfig {
  const ApiClientConfig({
    required this.environment,
    this.connectTimeout = const Duration(seconds: 15),
    this.receiveTimeout = const Duration(seconds: 30),
    this.sendTimeout = const Duration(seconds: 30),
    this.languageProvider,
    this.cityIdProvider,
    this.enableLogging = true,
  });

  final AppEnvironment environment;
  final Duration connectTimeout;
  final Duration receiveTimeout;
  final Duration sendTimeout;
  final String Function()? languageProvider;
  final String? Function()? cityIdProvider;
  final bool enableLogging;
}

/// Dio-backed REST client for NEXORA BFFs (`/v1`, CONSTITUTION §30).
class ApiClient {
  ApiClient._({
    required Dio dio,
    required ApiClientConfig config,
  })  : _dio = dio,
        _config = config;

  factory ApiClient.create({
    required ApiClientConfig config,
    SecureTokenStore? tokenStore,
    TokenRefresher? tokenRefresher,
    List<Interceptor> extraInterceptors = const [],
    Logger? logger,
  }) {
    final dio = Dio(
      BaseOptions(
        baseUrl: config.environment.baseUrl,
        connectTimeout: config.connectTimeout,
        receiveTimeout: config.receiveTimeout,
        sendTimeout: config.sendTimeout,
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json; charset=utf-8',
        },
        validateStatus: (status) => status != null && status < 500,
      ),
    );

    final log = logger ?? Logger('ApiClient');

    dio.interceptors.add(
      AcceptLanguageInterceptor(
        languageProvider: config.languageProvider ??
            () => config.environment.defaultLanguage,
      ),
    );

    if (config.cityIdProvider != null) {
      dio.interceptors.add(
        CityHeaderInterceptor(cityIdProvider: config.cityIdProvider!),
      );
    }

    dio.interceptors.add(IdempotencyInterceptor());

    if (config.enableLogging) {
      dio.interceptors.add(LoggingInterceptor(logger: log));
    }

    if (tokenStore != null && tokenRefresher != null) {
      dio.interceptors.add(
        AuthInterceptor(
          tokenStore: tokenStore,
          tokenRefresher: tokenRefresher,
          refreshDio: dio,
        ),
      );
    }

    dio.interceptors.addAll(extraInterceptors);

    CertificatePinningAdapter(
      pins: config.environment.certificatePins,
      logger: log,
    ).configure(dio);

    return ApiClient._(dio: dio, config: config);
  }

  final Dio _dio;
  final ApiClientConfig _config;

  Dio get dio => _dio;

  AppEnvironment get environment => _config.environment;

  Future<Result<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    T Function(dynamic json)? parser,
    Options? options,
  }) =>
      _request(
        () => _dio.get<dynamic>(
          path,
          queryParameters: queryParameters,
          options: options,
        ),
        parser: parser,
      );

  Future<Result<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    T Function(dynamic json)? parser,
    Options? options,
    String? idempotencyKey,
  }) =>
      _request(
        () => _dio.post<dynamic>(
          path,
          data: data,
          queryParameters: queryParameters,
          options: _withIdempotency(options, idempotencyKey),
        ),
        parser: parser,
      );

  Future<Result<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    T Function(dynamic json)? parser,
    Options? options,
    String? idempotencyKey,
  }) =>
      _request(
        () => _dio.put<dynamic>(
          path,
          data: data,
          queryParameters: queryParameters,
          options: _withIdempotency(options, idempotencyKey),
        ),
        parser: parser,
      );

  Future<Result<T>> patch<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    T Function(dynamic json)? parser,
    Options? options,
    String? idempotencyKey,
  }) =>
      _request(
        () => _dio.patch<dynamic>(
          path,
          data: data,
          queryParameters: queryParameters,
          options: _withIdempotency(options, idempotencyKey),
        ),
        parser: parser,
      );

  Future<Result<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    T Function(dynamic json)? parser,
    Options? options,
    String? idempotencyKey,
  }) =>
      _request(
        () => _dio.delete<dynamic>(
          path,
          data: data,
          queryParameters: queryParameters,
          options: _withIdempotency(options, idempotencyKey),
        ),
        parser: parser,
      );

  Options? _withIdempotency(Options? options, String? idempotencyKey) {
    if (idempotencyKey == null) {
      return options;
    }
    final merged = options ?? Options();
    merged.extra = {...?merged.extra, 'idempotency_key': idempotencyKey};
    return merged;
  }

  Future<Result<T>> _request<T>(
    Future<Response<dynamic>> Function() call, {
    T Function(dynamic json)? parser,
  }) async {
    try {
      final response = await call();
      final status = response.statusCode ?? 0;
      if (status >= 400) {
        throw DioException(
          requestOptions: response.requestOptions,
          response: response,
          type: DioExceptionType.badResponse,
        );
      }
      final data = response.data;
      if (parser == null) {
        return Success(data as T);
      }
      return Success(parser(data));
    } on DioException catch (error) {
      return Failure(DioErrorMapper.map(error));
    } on NexoraException catch (error) {
      return Failure(error);
    } catch (error) {
      return Failure(
        NexoraNetworkException(
          message: error.toString(),
          cause: error,
        ),
      );
    }
  }
}
