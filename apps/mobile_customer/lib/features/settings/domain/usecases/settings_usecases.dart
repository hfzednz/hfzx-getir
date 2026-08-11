import 'package:nexora_core/nexora_core.dart';

import '../entities/settings_entity.dart';
import '../repositories/settings_repository.dart';

class GetSettingsUseCase {
  const GetSettingsUseCase(this._repository);
  final SettingsRepository _repository;

  Future<Result<AppSettings>> call({String? id}) => _repository.fetch(id: id);
}

class ListSettingsUseCase {
  const ListSettingsUseCase(this._repository);
  final SettingsRepository _repository;

  Future<Result<List<AppSettings>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
