import 'package:nexora_core/nexora_core.dart';
import '../entities/report_kpis.dart';

abstract class ReportsRepository {
  Future<Result<ReportKpis>> fetchKpis();
}
