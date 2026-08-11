import 'package:nexora_core/nexora_core.dart';

import '../entities/about_entity.dart';
import '../repositories/about_repository.dart';

class GetAboutUseCase {
  const GetAboutUseCase(this._repository);
  final AboutRepository _repository;

  Future<Result<AboutInfo>> call({String? id}) => _repository.fetch(id: id);
}

class ListAboutUseCase {
  const ListAboutUseCase(this._repository);
  final AboutRepository _repository;

  Future<Result<List<AboutInfo>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
