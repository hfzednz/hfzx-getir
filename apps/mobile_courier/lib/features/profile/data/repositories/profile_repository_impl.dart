import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/profile_entity.dart';
import '../../domain/repositories/profile_repository.dart';
import '../datasources/profile_remote_datasource.dart';

class ProfileRepositoryImpl implements ProfileRepository {
  ProfileRepositoryImpl(this._remote);
  final ProfileRemoteDataSource _remote;

  @override
  Future<Result<CourierProfile>> fetch() => _remote.fetch();
}
