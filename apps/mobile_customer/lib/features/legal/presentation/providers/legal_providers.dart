import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/legal_remote_datasource.dart';
import '../../data/repositories/legal_repository_impl.dart';
import '../../domain/entities/legal_entity.dart';
import '../../domain/repositories/legal_repository.dart';
import '../../domain/usecases/legal_usecases.dart';

final legalRemoteDataSourceProvider = Provider<LegalRemoteDataSource>((ref) {
  return LegalRemoteDataSource(ref.watch(apiClientProvider));
});

final legalRepositoryProvider = Provider<LegalRepository>((ref) {
  return LegalRepositoryImpl(ref.watch(legalRemoteDataSourceProvider));
});

final getLegalUseCaseProvider = Provider((ref) =>
    GetLegalUseCase(ref.watch(legalRepositoryProvider)),);

final listLegalUseCaseProvider = Provider((ref) =>
    ListLegalUseCase(ref.watch(legalRepositoryProvider)),);

final legalListProvider = FutureProvider.autoDispose<List<LegalDocument>>((ref) async {
  final result = await ref.watch(listLegalUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});

final legalDocumentProvider =
    FutureProvider.autoDispose.family<LegalDocument, String>((ref, doc) async {
  final result = await ref.watch(getLegalUseCaseProvider).call(id: doc);
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (_) => LegalDocument(id: doc, title: '', payload: const {}),
  );
});
