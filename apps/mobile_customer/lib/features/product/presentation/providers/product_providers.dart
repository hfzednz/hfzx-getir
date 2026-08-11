import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../data/datasources/product_remote_datasource.dart';
import '../../data/repositories/product_repository_impl.dart';
import '../../domain/entities/product_entity.dart';
import '../../domain/repositories/product_repository.dart';
import '../../domain/usecases/product_usecases.dart';
import '../../../../di/providers.dart';

final productRemoteDataSourceProvider = Provider<ProductRemoteDataSource>((ref) {
  return ProductRemoteDataSource(ref.watch(apiClientProvider));
});

final productRepositoryProvider = Provider<ProductRepository>((ref) {
  return ProductRepositoryImpl(ref.watch(productRemoteDataSourceProvider));
});

final getProductUseCaseProvider = Provider(
  (ref) => GetProductUseCase(ref.watch(productRepositoryProvider)),
);

final toggleProductFavoriteUseCaseProvider = Provider(
  (ref) => ToggleProductFavoriteUseCase(ref.watch(productRepositoryProvider)),
);

final productDetailProvider = FutureProvider.family.autoDispose<Product, String>((ref, id) async {
  final result = await ref.watch(getProductUseCaseProvider).call(id: id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final productAlternativesProvider =
    FutureProvider.family.autoDispose<List<ProductSummary>, String>((ref, id) async {
  final result = await ref.watch(productRepositoryProvider).getAlternatives(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final productPriceHistoryProvider =
    FutureProvider.family.autoDispose<List<ProductPricePoint>, String>((ref, id) async {
  final result = await ref.watch(productRepositoryProvider).getPriceHistory(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
