import 'package:nexora_core/nexora_core.dart';

import '../entities/offer_entity.dart';

abstract class OffersRepository {
  Future<Result<List<Offer>>> listOffers();
  Future<Result<void>> accept(String offerId, {required String idempotencyKey});
  Future<Result<void>> reject(String offerId, {required String idempotencyKey});
}
