import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/offer_entity.dart';

class OffersRemoteDataSource {
  const OffersRemoteDataSource(this._client);
  final ApiClient _client;

  Future<Result<List<Offer>>> listOffers() {
    return _client.get<List<Offer>>(
      '/courier/offers',
      parser: (json) {
        final list = switch (json) {
          final List l => l,
          final Map m =>
            m['offers'] as List? ?? m['items'] as List? ?? const [],
          _ => const [],
        };
        return list
            .map((e) => Offer.fromJson(Map<String, dynamic>.from(e as Map)))
            .toList();
      },
    );
  }

  Future<Result<void>> accept(String id, {required String idempotencyKey}) {
    return _client.post<void>(
      '/courier/offers/$id/accept',
      idempotencyKey: idempotencyKey,
      parser: (_) {},
    );
  }

  Future<Result<void>> reject(String id, {required String idempotencyKey}) {
    return _client.post<void>(
      '/courier/offers/$id/reject',
      idempotencyKey: idempotencyKey,
      parser: (_) {},
    );
  }
}
