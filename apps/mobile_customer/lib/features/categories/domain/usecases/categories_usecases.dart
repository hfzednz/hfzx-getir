import 'package:nexora_core/nexora_core.dart';

import '../entities/categories_entity.dart';
import '../repositories/categories_repository.dart';

class GetCategoriesUseCase {
  const GetCategoriesUseCase(this._repository);
  final CategoriesRepository _repository;

  Future<Result<Category>> call({String? id}) => _repository.fetch(id: id);
}

class ListCategoriesUseCase {
  const ListCategoriesUseCase(this._repository);
  final CategoriesRepository _repository;

  Future<Result<List<Category>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
