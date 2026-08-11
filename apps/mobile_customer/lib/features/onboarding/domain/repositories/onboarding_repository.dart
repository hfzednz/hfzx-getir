import 'package:nexora_core/nexora_core.dart';

import '../entities/onboarding_entity.dart';

abstract class OnboardingRepository {
  Future<Result<OnboardingPage>> fetch({String? id});
  Future<Result<List<OnboardingPage>>> fetchList({Map<String, dynamic>? params});
  Future<Result<OnboardingPage>> mutate({required Map<String, dynamic> body, String? idempotencyKey});
}
