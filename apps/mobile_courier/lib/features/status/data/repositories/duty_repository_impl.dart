import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/duty_status.dart';
import '../../domain/repositories/duty_repository.dart';
import '../datasources/duty_remote_datasource.dart';

class DutyRepositoryImpl implements DutyRepository {
  DutyRepositoryImpl(this._remote);
  final DutyRemoteDataSource _remote;

  @override
  Future<Result<DutyStatus>> getStatus() => _remote.fetchStatus();

  @override
  Future<Result<DutyStatus>> setStatus(DutyStatus status) =>
      _remote.postStatus(status);
}
