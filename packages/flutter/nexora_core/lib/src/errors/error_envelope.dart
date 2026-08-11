import 'package:equatable/equatable.dart';

import 'nexora_error_code.dart';

/// Parsed public error envelope from NEXORA APIs (CONSTITUTION §16).
class NexoraErrorEnvelope extends Equatable {
  const NexoraErrorEnvelope({
    required this.code,
    required this.message,
    this.details,
    this.traceId,
    this.supportCode,
  });

  factory NexoraErrorEnvelope.fromJson(Map<String, dynamic> json) {
    final error = json['error'];
    if (error is! Map<String, dynamic>) {
      return const NexoraErrorEnvelope(
        code: NexoraErrorCode.unknown,
        message: 'Unexpected error response',
      );
    }

    return NexoraErrorEnvelope(
      code: NexoraErrorCode.fromCode(error['code'] as String?),
      message: (error['message'] as String?) ?? 'An error occurred',
      details: error['details'],
      traceId: error['trace_id'] as String?,
      supportCode: error['support_code'] as String?,
    );
  }

  final NexoraErrorCode code;
  final String message;
  final dynamic details;
  final String? traceId;
  final String? supportCode;

  @override
  List<Object?> get props => [code, message, details, traceId, supportCode];
}
