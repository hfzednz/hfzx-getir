import 'package:nexora_core/nexora_core.dart';

import '../entities/help_entity.dart';

abstract class HelpRepository {
  Future<Result<HelpArticle>> fetch({String? id});
  Future<Result<List<HelpArticle>>> fetchList({Map<String, dynamic>? params});
  Future<Result<HelpArticle>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
