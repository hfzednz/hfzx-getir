import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/settings_entity.dart';
import '../../domain/repositories/settings_repository.dart';
import '../datasources/settings_remote_datasource.dart';

class SettingsRepositoryImpl implements SettingsRepository {
  const SettingsRepositoryImpl(this._remote);
  final SettingsRemoteDataSource _remote;

  @override
  Future<Result<AppSettings>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<AppSettings>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<AppSettings>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
