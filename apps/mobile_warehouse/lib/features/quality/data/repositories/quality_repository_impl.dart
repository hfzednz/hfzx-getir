import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/qc_inspection.dart';
import '../../domain/repositories/quality_repository.dart';
import '../datasources/quality_remote_datasource.dart';

class QualityRepositoryImpl implements QualityRepository {
  QualityRepositoryImpl(this._remote);
  final QualityRemoteDataSource _remote;
  @override
  Future<Result<List<QcInspection>>> listQueue({String? stage}) => _remote.listQueue(stage: stage);
  @override
  Future<Result<QcInspection>> decide(String id, {required bool pass, String? notes, required String idempotencyKey}) =>
      _remote.decide(id, pass: pass, notes: notes, idempotencyKey: idempotencyKey);
}
