import 'package:nexora_core/nexora_core.dart';
import '../../domain/entities/shift_entity.dart';

class ShiftsRemoteDataSource {
  const ShiftsRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<WarehouseShift?>> current() {
    return _client.get<WarehouseShift?>(
      '/warehouse/shifts/current',
      parser: (json) {
        if (json == null) return null;
        return WarehouseShift.fromJson(Map<String, dynamic>.from(json as Map));
      },
    );
  }

  Future<Result<WarehouseShift>> _post(String path, String key) {
    return _client.post<WarehouseShift>(
      path,
      idempotencyKey: key,
      parser: (json) => WarehouseShift.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }

  Future<Result<WarehouseShift>> clockIn({required String idempotencyKey}) =>
      _post('/warehouse/shifts/clock-in', idempotencyKey);
  Future<Result<WarehouseShift>> clockOut({required String idempotencyKey}) =>
      _post('/warehouse/shifts/clock-out', idempotencyKey);
  Future<Result<WarehouseShift>> startBreak({required String idempotencyKey}) =>
      _post('/warehouse/shifts/break/start', idempotencyKey);
  Future<Result<WarehouseShift>> endBreak({required String idempotencyKey}) =>
      _post('/warehouse/shifts/break/end', idempotencyKey);

  Future<Result<List<WarehouseShift>>> attendance() {
    return _client.get<List<WarehouseShift>>(
      '/warehouse/shifts/attendance',
      parser: (json) {
        final list = json is List ? json : (json as Map)['items'] as List? ?? const [];
        return list.map((e) => WarehouseShift.fromJson(Map<String, dynamic>.from(e as Map))).toList();
      },
    );
  }
}
