import 'package:equatable/equatable.dart';

/// A queued offline mutation (CONSTITUTION §25).
class PendingMutation extends Equatable {
  const PendingMutation({
    required this.clientMutationId,
    required this.idempotencyKey,
    required this.method,
    required this.path,
    required this.body,
    required this.createdAt,
    this.retryCount = 0,
    this.nextAttemptAt,
    this.lastError,
  });

  factory PendingMutation.fromJson(Map<String, dynamic> json) {
    return PendingMutation(
      clientMutationId: json['client_mutation_id'] as String,
      idempotencyKey: json['idempotency_key'] as String,
      method: json['method'] as String,
      path: json['path'] as String,
      body: json['body'],
      createdAt: DateTime.parse(json['created_at'] as String),
      retryCount: json['retry_count'] as int? ?? 0,
      nextAttemptAt: json['next_attempt_at'] == null
          ? null
          : DateTime.parse(json['next_attempt_at'] as String),
      lastError: json['last_error'] as String?,
    );
  }

  final String clientMutationId;
  final String idempotencyKey;
  final String method;
  final String path;
  final dynamic body;
  final DateTime createdAt;
  final int retryCount;
  final DateTime? nextAttemptAt;
  final String? lastError;

  Map<String, dynamic> toJson() => {
        'client_mutation_id': clientMutationId,
        'idempotency_key': idempotencyKey,
        'method': method,
        'path': path,
        'body': body,
        'created_at': createdAt.toIso8601String(),
        'retry_count': retryCount,
        'next_attempt_at': nextAttemptAt?.toIso8601String(),
        'last_error': lastError,
      };

  PendingMutation copyWith({
    int? retryCount,
    DateTime? nextAttemptAt,
    String? lastError,
  }) {
    return PendingMutation(
      clientMutationId: clientMutationId,
      idempotencyKey: idempotencyKey,
      method: method,
      path: path,
      body: body,
      createdAt: createdAt,
      retryCount: retryCount ?? this.retryCount,
      nextAttemptAt: nextAttemptAt ?? this.nextAttemptAt,
      lastError: lastError ?? this.lastError,
    );
  }

  @override
  List<Object?> get props => [
        clientMutationId,
        idempotencyKey,
        method,
        path,
        body,
        createdAt,
        retryCount,
        nextAttemptAt,
        lastError,
      ];
}
