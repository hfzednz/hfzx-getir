import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/qc_inspection.dart';

class QualityRemoteDataSource {
  const QualityRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<QcInspection>>> listQueue({String? stage}) {
    return _client.get<List<QcInspection>>(
      '/warehouse/quality/queue',
      queryParameters: {if (stage != null) 'stage': stage},
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => QcInspection.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }

  Future<Result<QcInspection>> decide(String id, {required bool pass, String? notes, required String idempotencyKey}) {
    return _client.post<QcInspection>(
      '/warehouse/quality/$id/decide',
      data: {'pass': pass, if (notes != null) 'notes': notes},
      idempotencyKey: idempotencyKey,
      parser: (json) => QcInspection.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
