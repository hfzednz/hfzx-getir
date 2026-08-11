import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../di/providers.dart';
import '../../data/datasources/ai_remote_datasource.dart';
import '../../data/repositories/ai_repository_impl.dart';
import '../../domain/entities/ai_insights.dart';
import '../../domain/repositories/ai_repository.dart';

final aiRemoteDataSourceProvider = Provider((ref) => AiRemoteDataSource(ref.watch(apiClientProvider)));
final aiRepositoryProvider = Provider<AiRepository>((ref) => AiRepositoryImpl(ref.watch(aiRemoteDataSourceProvider)));
final aiHubProvider = FutureProvider.autoDispose<AiHubInsights>((ref) async {
  final r = await ref.watch(aiRepositoryProvider).fetchHub();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
