import 'package:nexora_core/nexora_core.dart';

import '../entities/city_entity.dart';
import '../repositories/city_repository.dart';

class GetCityUseCase {
  const GetCityUseCase(this._repository);
  final CityRepository _repository;

  Future<Result<CityContext>> call({String? id}) => _repository.fetch(id: id);
}

class ListCityUseCase {
  const ListCityUseCase(this._repository);
  final CityRepository _repository;

  Future<Result<List<CityContext>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
