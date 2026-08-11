import 'package:nexora_core/nexora_core.dart';

import '../entities/profile_entity.dart';

abstract class ProfileRepository {
  Future<Result<CourierProfile>> fetch();
}
