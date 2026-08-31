import 'package:nexora_core/nexora_core.dart';

import '../../../../shared/utils/json_list.dart';
import '../../domain/entities/addresses_entity.dart';
import '../models/addresses_model.dart';

class AddressesRemoteDataSource {
  const AddressesRemoteDataSource(this._client);
  final ApiClient _client;

  static const _listPath = '/addresses';

  Future<Result<Address>> fetch({String? id}) async {
    final path = id != null ? '$_listPath/$id' : _listPath;
    return _client.get<Address>(
      path,
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<List<Address>>> fetchList({Map<String, dynamic>? params}) async {
    return _client.get<List<Address>>(
      _listPath,
      queryParameters: params,
      parser: (json) => jsonObjectList(json)
          .map(AddressModel.fromJson)
          .map((e) => e.toEntity())
          .toList(),
    );
  }

  Future<Result<Address>> mutate({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Address>(
      _listPath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Address>> createAddress({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.post<Address>(
      _listPath,
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Address>> updateAddress({
    required String id,
    required Map<String, dynamic> body,
    String? idempotencyKey,
  }) async {
    return _client.patch<Address>(
      '$_listPath/$id',
      data: body,
      idempotencyKey: idempotencyKey,
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<void>> deleteAddress(String id) async {
    return _client.delete<void>(
      '$_listPath/$id',
      parser: (_) {},
    );
  }

  Future<Result<Address>> setDefault(String id) async {
    return _client.post<Address>(
      '$_listPath/$id/default',
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<Address>> setFavorite(String id, {required bool favorite}) async {
    return _client.post<Address>(
      '$_listPath/$id/favorite',
      data: {'favorite': favorite},
      parser: (json) =>
          AddressModel.fromJson(json as Map<String, dynamic>).toEntity(),
    );
  }

  Future<Result<AddressZoneValidation>> validateZone({
    required double lat,
    required double lng,
  }) async {
    return _client.get<AddressZoneValidation>(
      '$_listPath/validate-zone',
      queryParameters: {'lat': lat, 'lng': lng},
      parser: (json) =>
          AddressZoneValidation.fromJson(json as Map<String, dynamic>),
    );
  }
}
