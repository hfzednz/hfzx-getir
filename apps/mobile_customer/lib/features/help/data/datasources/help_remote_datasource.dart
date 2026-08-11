import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/help_entity.dart';
import '../models/help_model.dart';

class HelpRemoteDataSource {
  const HelpRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/help';
  static const _mutatePath = '/help';

  Future<Result<HelpArticle>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<HelpArticle>(
      path,
      parser: (json) => HelpArticleModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<HelpArticle>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<HelpArticle>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => HelpArticleModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<HelpArticle>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<HelpArticle>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => HelpArticleModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
