import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/help_remote_datasource.dart';
import '../../data/repositories/help_repository_impl.dart';
import '../../domain/entities/help_entity.dart';
import '../../domain/repositories/help_repository.dart';
import '../../domain/usecases/help_usecases.dart';

final helpRemoteDataSourceProvider = Provider<HelpRemoteDataSource>((ref) {
  return HelpRemoteDataSource(ref.watch(apiClientProvider));
});

final helpRepositoryProvider = Provider<HelpRepository>((ref) {
  return HelpRepositoryImpl(ref.watch(helpRemoteDataSourceProvider));
});

final getHelpUseCaseProvider = Provider((ref) =>
    GetHelpUseCase(ref.watch(helpRepositoryProvider)));

final listHelpUseCaseProvider = Provider((ref) =>
    ListHelpUseCase(ref.watch(helpRepositoryProvider)));

final helpListProvider = FutureProvider.autoDispose<List<HelpArticle>>((ref) async {
  final result = await ref.watch(listHelpUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
