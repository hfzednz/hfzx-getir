import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/map_remote_datasource.dart';
import '../../data/repositories/map_repository_impl.dart';
import '../../domain/entities/warehouse_layout.dart';
import '../../domain/repositories/map_repository.dart';

final mapRemoteDataSourceProvider = Provider((ref) => MapRemoteDataSource(ref.watch(apiClientProvider)));
final mapRepositoryProvider = Provider<MapRepository>((ref) => MapRepositoryImpl(ref.watch(mapRemoteDataSourceProvider)));
final warehouseMapProvider = FutureProvider.autoDispose<WarehouseLayout>((ref) async {
  final r = await ref.watch(mapRepositoryProvider).fetchLayout();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
