import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/datasources/search_remote_datasource.dart';
import '../../data/repositories/search_repository_impl.dart';
import '../../domain/entities/search_entity.dart';
import '../../domain/repositories/search_repository.dart';
import '../../domain/usecases/search_usecases.dart';
import '../../../cart/data/local/app_database.dart';
import '../../../product/domain/entities/product_entity.dart';
import '../../../stores/presentation/providers/stores_providers.dart';
import '../../../../di/providers.dart';

final searchRemoteDataSourceProvider = Provider<SearchRemoteDataSource>((ref) {
  return SearchRemoteDataSource(ref.watch(apiClientProvider));
});

final searchRepositoryProvider = Provider<SearchRepository>((ref) {
  return SearchRepositoryImpl(ref.watch(searchRemoteDataSourceProvider));
});

final semanticSearchUseCaseProvider = Provider(
  (ref) => SemanticSearchUseCase(ref.watch(searchRepositoryProvider)),
);

final searchSuggestionsUseCaseProvider = Provider(
  (ref) => SearchSuggestionsUseCase(ref.watch(searchRepositoryProvider)),
);

class SearchQuery {
  const SearchQuery({required this.text, this.filters = const SearchFilters()});

  final String text;
  final SearchFilters filters;
}

final searchFiltersProvider = StateProvider<SearchFilters>((ref) => const SearchFilters());

final catalogBrowseProvider =
    FutureProvider.autoDispose<List<ProductSummary>>((ref) async {
  final result = await ref.watch(semanticSearchUseCaseProvider).call('');
  return result.fold(onSuccess: (v) => v.items, onFailure: (e) => throw e);
});

final searchResultsProvider =
    FutureProvider.family.autoDispose<List<ProductSummary>, SearchQuery>((ref, query) async {
  if (query.text.trim().length < 2) return [];
  final storeId = ref.watch(selectedStoreIdProvider);
  final filters = query.filters.copyWith(storeId: storeId);
  final result = await ref.watch(semanticSearchUseCaseProvider).call(
        query.text,
        filters: filters,
      );
  return result.fold(
    onSuccess: (v) => v.items,
    onFailure: (e) => throw e,
  );
});

final searchSuggestionsProvider =
    FutureProvider.family.autoDispose<List<SearchSuggestion>, String>((ref, query) async {
  final result = await ref.watch(searchSuggestionsUseCaseProvider).call(query);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

/// Trending searches from suggestions API with empty query.
final trendingSearchesProvider =
    FutureProvider.autoDispose<List<SearchSuggestion>>((ref) async {
  final result = await ref.watch(searchSuggestionsUseCaseProvider).call('');
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final recentSearchesProvider = StreamProvider.autoDispose<List<RecentSearche>>((ref) {
  return ref.watch(databaseProvider).watchRecentSearches(limit: 10);
});

final barcodeSearchProvider =
    FutureProvider.family.autoDispose<List<ProductSummary>, String>((ref, barcode) async {
  final result = await ref.watch(searchRepositoryProvider).barcodeSearch(barcode);
  return result.fold(onSuccess: (v) => v.items, onFailure: (e) => throw e);
});
