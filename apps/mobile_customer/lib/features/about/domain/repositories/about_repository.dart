import 'package:nexora_core/nexora_core.dart';

import '../entities/about_entity.dart';

abstract class AboutRepository {
  Future<Result<AboutInfo>> fetch({String? id});
  Future<Result<List<AboutInfo>>> fetchList({Map<String, dynamic>? params});
  Future<Result<AboutInfo>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
