import 'package:nexora_core/nexora_core.dart';

import '../entities/settings_entity.dart';

abstract class SettingsRepository {
  Future<Result<AppSettings>> fetch({String? id});
  Future<Result<List<AppSettings>>> fetchList({Map<String, dynamic>? params});
  Future<Result<AppSettings>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
