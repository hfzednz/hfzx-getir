import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/addresses_remote_datasource.dart';
import '../../data/repositories/addresses_repository_impl.dart';
import '../../domain/entities/addresses_entity.dart';
import '../../domain/repositories/addresses_repository.dart';
import '../../domain/usecases/addresses_usecases.dart';

final addressesRemoteDataSourceProvider = Provider<AddressesRemoteDataSource>((ref) {
  return AddressesRemoteDataSource(ref.watch(apiClientProvider));
});

final addressesRepositoryProvider = Provider<AddressesRepository>((ref) {
  return AddressesRepositoryImpl(ref.watch(addressesRemoteDataSourceProvider));
});

final getAddressesUseCaseProvider = Provider(
  (ref) => GetAddressesUseCase(ref.watch(addressesRepositoryProvider)),
);

final listAddressesUseCaseProvider = Provider(
  (ref) => ListAddressesUseCase(ref.watch(addressesRepositoryProvider)),
);

final createAddressUseCaseProvider = Provider(
  (ref) => CreateAddressUseCase(ref.watch(addressesRepositoryProvider)),
);

final updateAddressUseCaseProvider = Provider(
  (ref) => UpdateAddressUseCase(ref.watch(addressesRepositoryProvider)),
);

final deleteAddressUseCaseProvider = Provider(
  (ref) => DeleteAddressUseCase(ref.watch(addressesRepositoryProvider)),
);

final setDefaultAddressUseCaseProvider = Provider(
  (ref) => SetDefaultAddressUseCase(ref.watch(addressesRepositoryProvider)),
);

final setFavoriteAddressUseCaseProvider = Provider(
  (ref) => SetFavoriteAddressUseCase(ref.watch(addressesRepositoryProvider)),
);

final validateAddressZoneUseCaseProvider = Provider(
  (ref) => ValidateAddressZoneUseCase(ref.watch(addressesRepositoryProvider)),
);

final addressesListProvider = FutureProvider.autoDispose<List<Address>>((ref) async {
  final result = await ref.watch(listAddressesUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});

final addressDetailProvider =
    FutureProvider.autoDispose.family<Address, String>((ref, id) async {
  final result = await ref.watch(getAddressesUseCaseProvider).call(id: id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
