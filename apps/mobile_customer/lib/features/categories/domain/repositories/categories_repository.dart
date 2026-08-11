import 'package:nexora_core/nexora_core.dart';

import '../entities/categories_entity.dart';

abstract class CategoriesRepository {
  Future<Result<Category>> fetch({String? id});
  Future<Result<List<Category>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Category>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
