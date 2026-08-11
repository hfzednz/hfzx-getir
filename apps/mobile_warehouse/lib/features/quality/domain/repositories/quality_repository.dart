import 'package:nexora_core/nexora_core.dart';
import '../entities/qc_inspection.dart';

abstract class QualityRepository {
  Future<Result<List<QcInspection>>> listQueue({String? stage});
  Future<Result<QcInspection>> decide(String id, {required bool pass, String? notes, required String idempotencyKey});
}
