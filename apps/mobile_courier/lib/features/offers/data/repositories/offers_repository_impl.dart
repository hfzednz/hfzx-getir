import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/offer_entity.dart';
import '../../domain/repositories/offers_repository.dart';
import '../datasources/offers_remote_datasource.dart';

class OffersRepositoryImpl implements OffersRepository {
  OffersRepositoryImpl(this._remote);
  final OffersRemoteDataSource _remote;

  @override
  Future<Result<List<Offer>>> listOffers() => _remote.listOffers();

  @override
  Future<Result<void>> accept(String offerId, {required String idempotencyKey}) =>
      _remote.accept(offerId, idempotencyKey: idempotencyKey);

  @override
  Future<Result<void>> reject(String offerId, {required String idempotencyKey}) =>
      _remote.reject(offerId, idempotencyKey: idempotencyKey);
}
