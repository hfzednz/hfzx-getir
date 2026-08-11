import 'package:nexora_core/nexora_core.dart';

import '../entities/search_entity.dart';

abstract class SearchRepository {
  Future<Result<SearchResult>> fetch({String? id});
  Future<Result<List<SearchResult>>> fetchList({Map<String, dynamic>? params});
  Future<Result<SearchResult>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<SearchResult>> semanticSearch(String query, {SearchFilters? filters});
  Future<Result<SearchResult>> barcodeSearch(String barcode);
  Future<Result<SearchResult>> imageSearch(List<int> imageBytes, {String? filename});
  Future<Result<List<SearchSuggestion>>> getSuggestions(String query);
}
