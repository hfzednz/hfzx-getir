import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/stores_remote_datasource.dart';
import '../../domain/entities/store_entity.dart';

final storesRemoteDataSourceProvider = Provider<StoresRemoteDataSource>((ref) {
  return StoresRemoteDataSource(ref.watch(apiClientProvider));
});

final storesListProvider = FutureProvider.autoDispose<List<StoreSummary>>((ref) async {
  final result = await ref.watch(storesRemoteDataSourceProvider).fetchList();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final storeDetailProvider =
    FutureProvider.autoDispose.family<StoreSummary, String>((ref, id) async {
  final result = await ref.watch(storesRemoteDataSourceProvider).fetch(id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
