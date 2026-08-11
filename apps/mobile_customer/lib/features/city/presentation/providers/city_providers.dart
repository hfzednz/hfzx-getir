import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/city_remote_datasource.dart';
import '../../data/repositories/city_repository_impl.dart';
import '../../domain/entities/city_entity.dart';
import '../../domain/repositories/city_repository.dart';
import '../../domain/usecases/city_usecases.dart';

final cityRemoteDataSourceProvider = Provider<CityRemoteDataSource>((ref) {
  return CityRemoteDataSource(ref.watch(apiClientProvider));
});

final cityRepositoryProvider = Provider<CityRepository>((ref) {
  return CityRepositoryImpl(ref.watch(cityRemoteDataSourceProvider));
});

final getCityUseCaseProvider = Provider((ref) =>
    GetCityUseCase(ref.watch(cityRepositoryProvider)));

final listCityUseCaseProvider = Provider((ref) =>
    ListCityUseCase(ref.watch(cityRepositoryProvider)));

final cityListProvider = FutureProvider.autoDispose<List<CityContext>>((ref) async {
  final result = await ref.watch(listCityUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
