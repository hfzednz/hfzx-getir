import 'package:dio/dio.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/delivery_job.dart';

class DeliveriesRemoteDataSource {
  const DeliveriesRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<DeliveryJob>>> listActive() {
    return _client.get<List<DeliveryJob>>(
      '/courier/deliveries',
      queryParameters: {'status': 'active'},
      parser: (json) {
        final list = switch (json) {
          final List l => l,
          final Map m =>
            m['deliveries'] as List? ?? m['items'] as List? ?? const [],
          _ => const [],
        };
        return list
            .map((e) =>
                DeliveryJob.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<DeliveryJob>> getById(String id) {
    return _client.get<DeliveryJob>(
      '/courier/deliveries/$id',
      parser: (json) =>
          DeliveryJob.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<DeliveryJob>> transition(
    String id,
    DeliveryLifecycleStatus status,
  ) {
    return _client.post<DeliveryJob>(
      '/courier/deliveries/$id/status',
      data: {'status': status.apiValue},
      parser: (json) =>
          DeliveryJob.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<DeliveryJob>> confirmPickup({
    required String id,
    required String handoffToken,
  }) {
    return _client.post<DeliveryJob>(
      '/courier/deliveries/$id/pickup',
      data: {'handoff_token': handoffToken},
      parser: (json) =>
          DeliveryJob.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<DeliveryJob>> submitPod({
    required String id,
    required String photoPath,
    String? otp,
    String? signatureNote,
  }) async {
    final form = FormData.fromMap({
      'photo': await MultipartFile.fromFile(photoPath),
      if (otp != null) 'otp': otp,
      if (signatureNote != null) 'signature_note': signatureNote,
    });
    return _client.post<DeliveryJob>(
      '/courier/deliveries/$id/pod',
      data: form,
      options: Options(contentType: 'multipart/form-data'),
      parser: (json) =>
          DeliveryJob.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<DeliveryJob>> markFailed({
    required String id,
    required String reasonCode,
    String? note,
  }) {
    return _client.post<DeliveryJob>(
      '/courier/deliveries/$id/fail',
      data: {
        'reason_code': reasonCode,
        if (note != null) 'note': note,
      },
      parser: (json) =>
          DeliveryJob.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
