import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/onboarding_entity.dart';
import '../models/onboarding_model.dart';

class OnboardingRemoteDataSource {
  const OnboardingRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/onboarding';
  static const _mutatePath = '/onboarding';

  Future<Result<OnboardingPage>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<OnboardingPage>(
      path,
      parser: (json) => OnboardingPageModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<OnboardingPage>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<OnboardingPage>>(
      _listPath,
      queryParameters: params,
      parser: (json) => (json as List<dynamic>)
          .map((e) => OnboardingPageModel.fromJson(e as Map<String, dynamic>).toEntity())
          .toList(),
    );
  }

  Future<Result<OnboardingPage>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<OnboardingPage>(
      _mutatePath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) => OnboardingPageModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }
}
