import 'package:nexora_core/nexora_core.dart';

import '../entities/search_entity.dart';
import '../repositories/search_repository.dart';

class GetSearchUseCase {
  const GetSearchUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<SearchResult>> call({String? id}) => _repository.fetch(id: id);
}

class ListSearchUseCase {
  const ListSearchUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<List<SearchResult>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class SemanticSearchUseCase {
  const SemanticSearchUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<SearchResult>> call(String query, {SearchFilters? filters}) =>
      _repository.semanticSearch(query, filters: filters);
}

class BarcodeSearchUseCase {
  const BarcodeSearchUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<SearchResult>> call(String barcode) => _repository.barcodeSearch(barcode);
}

class ImageSearchUseCase {
  const ImageSearchUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<SearchResult>> call(List<int> imageBytes, {String? filename}) =>
      _repository.imageSearch(imageBytes, filename: filename);
}

class SearchSuggestionsUseCase {
  const SearchSuggestionsUseCase(this._repository);
  final SearchRepository _repository;

  Future<Result<List<SearchSuggestion>>> call(String query) =>
      _repository.getSuggestions(query);
}
