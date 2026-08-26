import 'env.dart';

/// Resolved runtime environment for NEXORA mobile clients.
class AppEnvironment {
  const AppEnvironment({
    required this.name,
    required this.baseUrl,
    required this.wsUrl,
    this.defaultLanguage = Env.defaultLanguage,
    this.certificatePins = const [],
  });

  /// Build from compile-time [Env] constants.
  factory AppEnvironment.fromEnv() {
    final pins = Env.certificatePins
        .split(',')
        .map((pin) => pin.trim())
        .where((pin) => pin.isNotEmpty)
        .toList(growable: false);

    var name = Env.name;
    var baseUrl = Env.baseUrl;
    var wsUrl = Env.wsUrl;
    if (name == 'prod' && baseUrl.contains('dev.nexora.local')) {
      baseUrl = 'https://api.nexora.io/v1';
      wsUrl = 'wss://realtime.nexora.io/v1';
    } else if (name == 'staging' && baseUrl.contains('dev.nexora.local')) {
      baseUrl = 'https://api.staging.nexora.io/v1';
      wsUrl = 'wss://realtime.staging.nexora.io/v1';
    }

    return AppEnvironment(
      name: name,
      baseUrl: baseUrl,
      wsUrl: wsUrl,
      defaultLanguage: Env.defaultLanguage,
      certificatePins: pins,
    );
  }

  AppEnvironment copyWith({
    String? name,
    String? baseUrl,
    String? wsUrl,
    String? defaultLanguage,
    List<String>? certificatePins,
  }) {
    return AppEnvironment(
      name: name ?? this.name,
      baseUrl: baseUrl ?? this.baseUrl,
      wsUrl: wsUrl ?? this.wsUrl,
      defaultLanguage: defaultLanguage ?? this.defaultLanguage,
      certificatePins: certificatePins ?? this.certificatePins,
    );
  }

  /// Customer BFF is mounted at `/v1/customer`.
  AppEnvironment get forCustomerBff {
    final trimmed = baseUrl.replaceAll(RegExp(r'/+$'), '');
    if (trimmed.endsWith('/v1/customer')) return this;
    if (trimmed.endsWith('/v1')) {
      return copyWith(baseUrl: '$trimmed/customer');
    }
    return this;
  }

  /// Named presets for tests and local overrides.
  const AppEnvironment.dev()
      : name = 'dev',
        baseUrl = 'https://api.dev.nexora.local/v1',
        wsUrl = 'wss://realtime.dev.nexora.local/v1',
        defaultLanguage = 'en',
        certificatePins = const [];

  const AppEnvironment.staging()
      : name = 'staging',
        baseUrl = 'https://api.staging.nexora.io/v1',
        wsUrl = 'wss://realtime.staging.nexora.io/v1',
        defaultLanguage = 'en',
        certificatePins = const [];

  const AppEnvironment.prod()
      : name = 'prod',
        baseUrl = 'https://api.nexora.io/v1',
        wsUrl = 'wss://realtime.nexora.io/v1',
        defaultLanguage = 'en',
        certificatePins = const [];

  final String name;
  final String baseUrl;
  final String wsUrl;
  final String defaultLanguage;
  final List<String> certificatePins;

  bool get isProduction => name == 'prod';

  @override
  String toString() => 'AppEnvironment($name, $baseUrl)';
}
