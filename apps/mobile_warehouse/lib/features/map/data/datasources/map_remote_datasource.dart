import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/warehouse_layout.dart';

class MapRemoteDataSource {
  const MapRemoteDataSource(this._client);
  final ApiClient _client;
  Future<Result<WarehouseLayout>> fetchLayout() {
    return _client.get<WarehouseLayout>(
      '/warehouse/map',
      parser: (json) => WarehouseLayout.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
