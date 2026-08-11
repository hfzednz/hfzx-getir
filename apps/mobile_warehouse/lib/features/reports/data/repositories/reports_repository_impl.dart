import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/report_kpis.dart';
import '../../domain/repositories/reports_repository.dart';
import '../datasources/reports_remote_datasource.dart';

class ReportsRepositoryImpl implements ReportsRepository {
  ReportsRepositoryImpl(this._remote);
  final ReportsRemoteDataSource _remote;
  @override
  Future<Result<ReportKpis>> fetchKpis() => _remote.fetchKpis();
}
