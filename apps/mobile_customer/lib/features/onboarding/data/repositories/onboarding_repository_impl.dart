import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/onboarding_entity.dart';
import '../../domain/repositories/onboarding_repository.dart';
import '../datasources/onboarding_remote_datasource.dart';

class OnboardingRepositoryImpl implements OnboardingRepository {
  const OnboardingRepositoryImpl(this._remote);
  final OnboardingRemoteDataSource _remote;

  @override
  Future<Result<OnboardingPage>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<OnboardingPage>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<OnboardingPage>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);
}
