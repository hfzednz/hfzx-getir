import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/shift_entity.dart';

class ShiftsRemoteDataSource {
  const ShiftsRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<CourierShift>>> listShifts() {
    return _client.get<List<CourierShift>>(
      '/courier/shifts',
      parser: (json) {
        final list = switch (json) {
          final List l => l,
          final Map m =>
            m['shifts'] as List? ?? m['items'] as List? ?? const [],
          _ => const [],
        };
        return list
            .map((e) =>
                CourierShift.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<CourierShift>> postAction(String path, {String? id}) {
    final url = id == null ? '/courier/shifts$path' : '/courier/shifts/$id$path';
    return _client.post<CourierShift>(
      url,
      parser: (json) =>
          CourierShift.fromJson(Map<String, dynamic>.from(json as Map)),
    );
  }
}
