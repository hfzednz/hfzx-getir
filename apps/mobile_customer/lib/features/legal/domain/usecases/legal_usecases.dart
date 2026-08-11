import 'package:nexora_core/nexora_core.dart';

import '../entities/legal_entity.dart';
import '../repositories/legal_repository.dart';

class GetLegalUseCase {
  const GetLegalUseCase(this._repository);
  final LegalRepository _repository;

  Future<Result<LegalDocument>> call({String? id}) => _repository.fetch(id: id);
}

class ListLegalUseCase {
  const ListLegalUseCase(this._repository);
  final LegalRepository _repository;

  Future<Result<List<LegalDocument>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
