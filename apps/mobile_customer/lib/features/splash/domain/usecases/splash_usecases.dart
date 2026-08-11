import 'package:nexora_core/nexora_core.dart';

import '../entities/splash_entity.dart';
import '../repositories/splash_repository.dart';

class GetSplashUseCase {
  const GetSplashUseCase(this._repository);
  final SplashRepository _repository;

  Future<Result<SplashState>> call({String? id}) => _repository.fetch(id: id);
}

class ListSplashUseCase {
  const ListSplashUseCase(this._repository);
  final SplashRepository _repository;

  Future<Result<List<SplashState>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
