import 'package:nexora_core/nexora_core.dart';

import '../entities/splash_entity.dart';

abstract class SplashRepository {
  Future<Result<SplashState>> fetch({String? id});
  Future<Result<List<SplashState>>> fetchList({Map<String, dynamic>? params});
  Future<Result<SplashState>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
