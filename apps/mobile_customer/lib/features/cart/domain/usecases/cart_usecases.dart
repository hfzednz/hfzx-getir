import 'package:nexora_core/nexora_core.dart';

import '../entities/cart_entity.dart';
import '../repositories/cart_repository.dart';

class GetCartUseCase {
  const GetCartUseCase(this._repository);
  final CartRepository _repository;

  Future<Result<Cart>> call({String? id}) => _repository.fetch(id: id);
}

class ListCartUseCase {
  const ListCartUseCase(this._repository);
  final CartRepository _repository;

  Future<Result<List<Cart>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
