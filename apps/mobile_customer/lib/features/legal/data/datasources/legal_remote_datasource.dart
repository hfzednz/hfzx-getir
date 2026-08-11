import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/legal_entity.dart';
import '../models/legal_model.dart';

class LegalRemoteDataSource {
  const LegalRemoteDataSource(this._client);
  final ApiClient _client;

  static const _basePath = '/legal';
  static const _mutatePath = '/legal';

  Future<Result<LegalDocument>> fetch({String? id}) async {
    final path = id != null ? '$_basePath/$id' : _basePath;
    return _client.get<LegalDocument>(
      path,
      parser: (json) => LegalDocumentModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<LegalDocument>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<LegalDocument>>(
      _basePath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => LegalDocumentModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<LegalDocument>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<LegalDocument>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => LegalDocumentModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
