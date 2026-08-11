import 'package:nexora_core/nexora_core.dart';

import '../entities/onboarding_entity.dart';
import '../repositories/onboarding_repository.dart';

class GetOnboardingUseCase {
  const GetOnboardingUseCase(this._repository);
  final OnboardingRepository _repository;

  Future<Result<OnboardingPage>> call({String? id}) => _repository.fetch(id: id);
}

class ListOnboardingUseCase {
  const ListOnboardingUseCase(this._repository);
  final OnboardingRepository _repository;

  Future<Result<List<OnboardingPage>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}
