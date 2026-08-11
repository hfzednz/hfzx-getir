import 'package:nexora_core/nexora_core.dart';

import '../entities/home_entity.dart';
import '../repositories/home_repository.dart';

class GetHomeUseCase {
  const GetHomeUseCase(this._repository);
  final HomeRepository _repository;

  Future<Result<HomeFeed>> call({String? id}) => _repository.fetch(id: id);
}

class ListHomeUseCase {
  const ListHomeUseCase(this._repository);
  final HomeRepository _repository;

  Future<Result<List<HomeFeed>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
