import 'package:nexora_core/nexora_core.dart';

import '../../features/addresses/domain/entities/addresses_entity.dart';

/// Delivery address validation for checkout and address CRUD.
abstract final class AddressValidator {
  static const minTitleLength = 2;
  static const minFormattedLength = 5;

  static String? validateTitle(String? value) {
    final title = value?.trim() ?? '';
    if (title.isEmpty) return 'Address label is required';
    if (title.length < minTitleLength) {
      return 'Address label must be at least $minTitleLength characters';
    }
    return null;
  }

  static String? validateFormatted(String? value) {
    final line = value?.trim() ?? '';
    if (line.isEmpty) return 'Street address is required';
    if (line.length < minFormattedLength) {
      return 'Enter a complete street address';
    }
    return null;
  }

  static String? validateCoordinates({double? lat, double? lng}) {
    final result = validateCoordinatesResult(lat: lat, lng: lng);
    if (result.isSuccess) return null;
    return result.errorOrNull?.message;
  }

  static Result<({double lat, double lng})> validateCoordinatesResult({
    double? lat,
    double? lng,
  }) {
    if (lat == null || lng == null) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Location coordinates are required',
          details: {'field': 'coordinates'},
        ),
      );
    }

    if (lat < -90 || lat > 90 || lng < -180 || lng > 180) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Location coordinates are out of range',
          details: {'field': 'coordinates', 'lat': lat, 'lng': lng},
        ),
      );
    }

    return Success((lat: lat, lng: lng));
  }

  static Result<Address> validateForCheckout(
    Address? address, {
    required bool serviceable,
  }) {
    if (address == null || address.id.isEmpty) {
      return const Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery address is required',
          details: {'field': 'address_id'},
        ),
      );
    }

    if (validateTitle(address.title) != null || validateFormatted(address.formatted) != null) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery address is incomplete',
          details: {'field': 'address', 'address_id': address.id},
        ),
      );
    }

    final coords = validateCoordinatesResult(lat: address.lat, lng: address.lng);
    if (coords.isFailure) {
      return Failure(coords.errorOrNull!);
    }

    if (!serviceable) {
      return Failure(
        NexoraValidationException(
          code: NexoraErrorCode.validationFailed,
          message: 'Delivery is not available at this address',
          details: {'field': 'address_id', 'address_id': address.id},
        ),
      );
    }

    return Success(address);
  }
}
