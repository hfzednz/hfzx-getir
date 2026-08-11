import 'package:nexora_core/nexora_core.dart';
import '../entities/ai_insights.dart';

abstract class AiRepository {
  Future<Result<AiHubInsights>> fetchHub();
}
