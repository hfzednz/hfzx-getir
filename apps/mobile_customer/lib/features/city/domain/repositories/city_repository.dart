import 'package:nexora_core/nexora_core.dart';

import '../entities/city_entity.dart';

abstract class CityRepository {
  Future<Result<CityContext>> fetch({String? id});
  Future<Result<List<CityContext>>> fetchList({Map<String, dynamic>? params});
  Future<Result<CityContext>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
