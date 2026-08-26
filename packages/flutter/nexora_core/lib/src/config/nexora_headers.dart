/// Standard NEXORA HTTP header names (CONSTITUTION §30).
abstract final class NexoraHeaders {
  static const authorization = 'Authorization';
  static const idempotencyKey = 'Idempotency-Key';
  static const acceptLanguage = 'Accept-Language';
  static const cityId = 'X-Nexora-City-Id';
  static const tenantId = 'X-Tenant-Id';
  static const requestId = 'X-Request-Id';
  static const contentType = 'Content-Type';
  static const accept = 'Accept';
}
