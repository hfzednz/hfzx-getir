import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/datasources/categories_remote_datasource.dart';
import '../../data/repositories/categories_repository_impl.dart';
import '../../domain/entities/categories_entity.dart';
import '../../../home/domain/entities/home_entity.dart';
import '../../domain/repositories/categories_repository.dart';
import '../../domain/usecases/categories_usecases.dart';
import '../../../../di/providers.dart';

final categoriesRemoteDataSourceProvider = Provider<CategoriesRemoteDataSource>((ref) {
  return CategoriesRemoteDataSource(ref.watch(apiClientProvider));
});

final categoriesRepositoryProvider = Provider<CategoriesRepository>((ref) {
  return CategoriesRepositoryImpl(ref.watch(categoriesRemoteDataSourceProvider));
});

final listCategoriesUseCaseProvider = Provider(
  (ref) => ListCategoriesUseCase(ref.watch(categoriesRepositoryProvider)),
);

final categoryFiltersProvider = StateProvider<CategoryFilter>((ref) => const CategoryFilter());

final categoriesListProvider = FutureProvider.autoDispose<List<Category>>((ref) async {
  final result = await ref.watch(listCategoriesUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final categoryProductsProvider = FutureProvider.family.autoDispose<List<HomeProduct>, String>((ref, id) async {
  final client = ref.watch(apiClientProvider);
  final filters = ref.watch(categoryFiltersProvider);
  final result = await client.get<List<HomeProduct>>(
    '/categories/$id/products',
    queryParameters: filters.toQueryParams(),
    parser: (json) => (json['items'] as List<dynamic>? ?? [])
        .map((e) => HomeProduct.fromJson(e as Map<String, dynamic>))
        .toList(),
  );
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
