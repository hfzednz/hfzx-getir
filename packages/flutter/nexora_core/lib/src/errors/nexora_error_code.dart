/// Canonical NEXORA public error codes (CONSTITUTION §16).
enum NexoraErrorCode {
  // Validation / client
  validationFailed('VALIDATION_FAILED'),
  invalidRequest('INVALID_REQUEST'),

  // Auth
  authRequired('AUTH_REQUIRED'),
  authInvalid('AUTH_INVALID'),
  authExpired('AUTH_EXPIRED'),
  authForbidden('AUTH_FORBIDDEN'),
  refreshTokenReused('REFRESH_TOKEN_REUSED'),

  // Resource
  notFound('NOT_FOUND'),
  conflict('CONFLICT'),
  idempotencyReplay('IDEMPOTENCY_REPLAY'),

  // Rate / availability
  rateLimited('RATE_LIMITED'),
  serviceUnavailable('SERVICE_UNAVAILABLE'),

  // Server
  internalError('INTERNAL_ERROR'),
  unknown('UNKNOWN'),

  // Network (client-side)
  networkError('NETWORK_ERROR'),
  timeout('TIMEOUT'),
  cancelled('CANCELLED');

  const NexoraErrorCode(this.code);

  final String code;

  static NexoraErrorCode fromCode(String? raw) {
    if (raw == null || raw.isEmpty) {
      return NexoraErrorCode.unknown;
    }
    for (final value in NexoraErrorCode.values) {
      if (value.code == raw) {
        return value;
      }
    }
    return NexoraErrorCode.unknown;
  }
}
