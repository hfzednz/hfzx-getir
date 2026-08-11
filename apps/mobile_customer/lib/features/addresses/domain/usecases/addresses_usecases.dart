import 'package:nexora_core/nexora_core.dart';

import '../entities/addresses_entity.dart';
import '../repositories/addresses_repository.dart';

class GetAddressesUseCase {
  const GetAddressesUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<Address>> call({String? id}) => _repository.fetch(id: id);
}

class ListAddressesUseCase {
  const ListAddressesUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<List<Address>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class CreateAddressUseCase {
  const CreateAddressUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<Address>> call({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _repository.createAddress(body: body, idempotencyKey: idempotencyKey);
}

class UpdateAddressUseCase {
  const UpdateAddressUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<Address>> call({
    required String id,
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) =>
      _repository.updateAddress(id: id, body: body, idempotencyKey: idempotencyKey);
}

class DeleteAddressUseCase {
  const DeleteAddressUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<void>> call(String id) => _repository.deleteAddress(id);
}

class SetDefaultAddressUseCase {
  const SetDefaultAddressUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<Address>> call(String id) => _repository.setDefault(id);
}

class SetFavoriteAddressUseCase {
  const SetFavoriteAddressUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<Address>> call(String id, {required bool favorite}) =>
      _repository.setFavorite(id, favorite: favorite);
}

class ValidateAddressZoneUseCase {
  const ValidateAddressZoneUseCase(this._repository);
  final AddressesRepository _repository;

  Future<Result<AddressZoneValidation>> call({
    required double lat,
    required double lng,
  }) =>
      _repository.validateZone(lat: lat, lng: lng);
}
