import 'package:nexora_core/nexora_core.dart';

import '../entities/legal_entity.dart';

abstract class LegalRepository {
  Future<Result<LegalDocument>> fetch({String? id});
  Future<Result<List<LegalDocument>>> fetchList({Map<String, dynamic>? params});
  Future<Result<LegalDocument>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
