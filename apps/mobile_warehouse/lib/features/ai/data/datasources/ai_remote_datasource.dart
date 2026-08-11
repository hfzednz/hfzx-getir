import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/ai_insights.dart';

class AiRemoteDataSource {
  const AiRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<AiHubInsights>> fetchHub() {
    return _client.get<AiHubInsights>(
      '/warehouse/ai/hub',
      parser: (json) => AiHubInsights.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
