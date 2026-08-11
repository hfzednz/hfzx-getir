import 'package:dio/dio.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/search_entity.dart';
import '../models/search_model.dart';

class SearchRemoteDataSource {
  const SearchRemoteDataSource(this._client);
  final ApiClient _client;

  static const _basePath = '/search';

  SearchResult _parseResult(dynamic json) =>
      SearchResultModel.fromJson(json as Map<String, dynamic>).toEntity();

  Future<Result<SearchResult>> fetch({String? id}) async {
    final path = id != null ? '$_basePath/$id' : _basePath;
    return _client.get<SearchResult>(path, parser: (json) => _parseResult(json));
  }

  Future<Result<List<SearchResult>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<SearchResult>>(
      _basePath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => SearchResultModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<SearchResult>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<SearchResult>(
      _basePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => _parseResult(json),
    );
  }

  Future<Result<SearchResult>> semanticSearch(String query, {SearchFilters? filters}) async {
    return _client.get<SearchResult>(
      '$_basePath/semantic',
      queryParameters: {'q': query, ...?filters?.toQueryParams()},
      parser: (json) => _parseResult(json),
    );
  }

  Future<Result<SearchResult>> barcodeSearch(String barcode) async {
    return _client.get<SearchResult>(
      '$_basePath/barcode',
      queryParameters: {'code': barcode},
      parser: (json) => _parseResult(json),
    );
  }

  Future<Result<SearchResult>> imageSearch(List<int> imageBytes, {String? filename}) async {
    final formData = FormData.fromMap({
      'image': MultipartFile.fromBytes(
        imageBytes,
        filename: filename ?? 'search.jpg',
      ),
    });
    return _client.post<SearchResult>(
      '$_basePath/image',
      data: formData,
      options: Options(contentType: 'multipart/form-data'),
      parser: (json) => _parseResult(json),
    );
  }

  Future<Result<List<SearchSuggestion>>> getSuggestions(String query) async {
    return _client.get<List<SearchSuggestion>>(
      '$_basePath/suggestions',
      queryParameters: {'q': query},
      parser: (json) => (json['suggestions'] as List<dynamic>? ?? json as List<dynamic>)
          .map((e) => SearchSuggestion.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }
}
