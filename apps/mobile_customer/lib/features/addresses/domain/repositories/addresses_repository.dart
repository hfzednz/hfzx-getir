import 'package:nexora_core/nexora_core.dart';

import '../entities/addresses_entity.dart';

abstract class AddressesRepository {
  Future<Result<Address>> fetch({String? id});
  Future<Result<List<Address>>> fetchList({Map<String, dynamic>? params});
  Future<Result<Address>> mutate({required Map<String, dynamic> body, String? idempotencyKey});

  Future<Result<Address>> createAddress({
    required Map<String, dynamic> body,
    String? idempotencyKey,
  });

  Future<Result<Address>> updateAddress({
    required String id,
    required Map<String, dynamic> body,
    String? idempotencyKey,
  });

  Future<Result<void>> deleteAddress(String id);

  Future<Result<Address>> setDefault(String id);

  Future<Result<Address>> setFavorite(String id, {required bool favorite});

  Future<Result<AddressZoneValidation>> validateZone({
    required double lat,
    required double lng,
  });
}
