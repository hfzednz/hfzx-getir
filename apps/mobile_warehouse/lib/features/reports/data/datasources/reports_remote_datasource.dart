import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/report_kpis.dart';

class ReportsRemoteDataSource {
  const ReportsRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<ReportKpis>> fetchKpis() {
    return _client.get<ReportKpis>(
      '/warehouse/reports/kpis',
      parser: (json) => ReportKpis.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
