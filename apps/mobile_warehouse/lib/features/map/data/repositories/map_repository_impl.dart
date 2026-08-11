import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/warehouse_layout.dart';
import '../../domain/repositories/map_repository.dart';
import '../datasources/map_remote_datasource.dart';

class MapRepositoryImpl implements MapRepository {
  MapRepositoryImpl(this._remote);
  final MapRemoteDataSource _remote;
  @override
  Future<Result<WarehouseLayout>> fetchLayout() => _remote.fetchLayout();
}
