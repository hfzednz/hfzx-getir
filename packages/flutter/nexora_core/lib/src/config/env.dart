/// Compile-time environment constants (Dart `--dart-define` / `--dart-define-from-file`).
abstract final class Env {
  static const name = String.fromEnvironment(
    'NEXORA_ENV',
    defaultValue: 'dev',
  );

  static const baseUrl = String.fromEnvironment(
    'NEXORA_BASE_URL',
    defaultValue: 'https://api.dev.nexora.local/v1',
  );

  static const wsUrl = String.fromEnvironment(
    'NEXORA_WS_URL',
    defaultValue: 'wss://realtime.dev.nexora.local/v1',
  );

  static const defaultLanguage = String.fromEnvironment(
    'NEXORA_DEFAULT_LANGUAGE',
    defaultValue: 'en',
  );

  static const certificatePins = String.fromEnvironment(
    'NEXORA_CERTIFICATE_PINS',
    defaultValue: '',
  );
}
