import 'package:nexora_core/nexora_core.dart';

import '../entities/help_entity.dart';
import '../repositories/help_repository.dart';

class GetHelpUseCase {
  const GetHelpUseCase(this._repository);
  final HelpRepository _repository;

  Future<Result<HelpArticle>> call({String? id}) => _repository.fetch(id: id);
}

class ListHelpUseCase {
  const ListHelpUseCase(this._repository);
  final HelpRepository _repository;

  Future<Result<List<HelpArticle>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
