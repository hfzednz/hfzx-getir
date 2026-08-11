import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/search_entity.dart';
import '../../domain/repositories/search_repository.dart';
import '../datasources/search_remote_datasource.dart';

class SearchRepositoryImpl implements SearchRepository {
  const SearchRepositoryImpl(this._remote);
  final SearchRemoteDataSource _remote;

  @override
  Future<Result<SearchResult>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<SearchResult>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<SearchResult>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<SearchResult>> semanticSearch(String query, {SearchFilters? filters}) =>
      _remote.semanticSearch(query, filters: filters);

  @override
  Future<Result<SearchResult>> barcodeSearch(String barcode) =>
      _remote.barcodeSearch(barcode);

  @override
  Future<Result<SearchResult>> imageSearch(List<int> imageBytes, {String? filename}) =>
      _remote.imageSearch(imageBytes, filename: filename);

  @override
  Future<Result<List<SearchSuggestion>>> getSuggestions(String query) =>
      _remote.getSuggestions(query);
}
