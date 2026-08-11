import 'package:nexora_core/nexora_core.dart';

import '../../domain/entities/addresses_entity.dart';
import '../../domain/repositories/addresses_repository.dart';
import '../datasources/addresses_remote_datasource.dart';

class AddressesRepositoryImpl implements AddressesRepository {
  const AddressesRepositoryImpl(this._remote);
  final AddressesRemoteDataSource _remote;

  @override
  Future<Result<Address>> fetch({String? id}) => _remote.fetch(id: id);

  @override
  Future<Result<List<Address>>> fetchList({Map<String, dynamic>? params}) =>
      _remote.fetchList(params: params);

  @override
  Future<Result<Address>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.mutate(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Address>> createAddress({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.createAddress(body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<Address>> updateAddress({
    required String id,
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _remote.updateAddress(id: id, body: body, idempotencyKey: idempotencyKey);

  @override
  Future<Result<void>> deleteAddress(String id) => _remote.deleteAddress(id);

  @override
  Future<Result<Address>> setDefault(String id) => _remote.setDefault(id);

  @override
  Future<Result<Address>> setFavorite(String id, {required bool favorite}) =>
      _remote.setFavorite(id, favorite: favorite);

  @override
  Future<Result<AddressZoneValidation>> validateZone({
    required double lat,
    required double lng,
  }) =>
      _remote.validateZone(lat: lat, lng: lng);
}
