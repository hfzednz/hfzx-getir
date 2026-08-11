import 'package:nexora_core/nexora_core.dart';

import '../entities/home_entity.dart';

abstract class HomeRepository {
  Future<Result<HomeFeed>> fetch({String? id});
  Future<Result<List<HomeFeed>>> fetchList({Map<String, dynamic>? params});
  Future<Result<HomeFeed>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
