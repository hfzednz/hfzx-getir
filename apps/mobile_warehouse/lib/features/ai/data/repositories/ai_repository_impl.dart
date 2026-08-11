import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/ai_insights.dart';
import '../../domain/repositories/ai_repository.dart';
import '../datasources/ai_remote_datasource.dart';

class AiRepositoryImpl implements AiRepository {
  AiRepositoryImpl(this._remote);
  final AiRemoteDataSource _remote;
  @override
  Future<Result<AiHubInsights>> fetchHub() => _remote.fetchHub();
}
