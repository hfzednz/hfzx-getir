import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/checkout_rules.dart';
import '../../../addresses/domain/entities/addresses_entity.dart';
import '../../../addresses/presentation/providers/addresses_providers.dart';
import '../../data/datasources/checkout_remote_datasource.dart';
import '../../data/datasources/payment_methods_remote_datasource.dart';
import '../../data/repositories/checkout_repository_impl.dart';
import '../../domain/entities/checkout_entity.dart';
import '../../domain/repositories/checkout_repository.dart';
import '../../domain/usecases/checkout_usecases.dart';
import 'checkout_controller.dart';

export 'checkout_controller.dart';

final checkoutRemoteDataSourceProvider = Provider<CheckoutRemoteDataSource>((ref) {
  return CheckoutRemoteDataSource(ref.watch(apiClientProvider));
});

final checkoutRepositoryProvider = Provider<CheckoutRepository>((ref) {
  return CheckoutRepositoryImpl(ref.watch(checkoutRemoteDataSourceProvider));
});

final paymentMethodsRemoteDataSourceProvider =
    Provider<PaymentMethodsRemoteDataSource>((ref) {
  return PaymentMethodsRemoteDataSource(ref.watch(apiClientProvider));
});

final paymentMethodsRepositoryProvider = Provider<PaymentMethodsRepository>((ref) {
  return PaymentMethodsRepositoryImpl(ref.watch(paymentMethodsRemoteDataSourceProvider));
});

final getCheckoutUseCaseProvider =
    Provider((ref) => GetCheckoutUseCase(ref.watch(checkoutRepositoryProvider)));

final listCheckoutUseCaseProvider =
    Provider((ref) => ListCheckoutUseCase(ref.watch(checkoutRepositoryProvider)));

final getCheckoutQuoteUseCaseProvider =
    Provider((ref) => GetCheckoutQuoteUseCase(ref.watch(checkoutRepositoryProvider)));

final confirmCheckoutUseCaseProvider =
    Provider((ref) => ConfirmCheckoutUseCase(ref.watch(checkoutRepositoryProvider)));

final verifyCheckoutQuoteUseCaseProvider =
    Provider((ref) => VerifyCheckoutQuoteUseCase(ref.watch(checkoutRepositoryProvider)));

final checkoutControllerProvider =
    NotifierProvider<CheckoutController, CheckoutState>(CheckoutController.new);

/// Derives a [CheckoutDraft] from controller state + resolved address.
final checkoutDraftProvider = Provider<CheckoutDraft>((ref) {
  final state = ref.watch(checkoutControllerProvider);
  final addresses = ref.watch(addressesListProvider).maybeWhen(
        data: (list) => list,
        orElse: () => const <Address>[],
      );
  Address? address;
  if (state.addressId != null) {
    for (final a in addresses) {
      if (a.id == state.addressId) {
        address = a;
        break;
      }
    }
  }

  return CheckoutDraft(
    address: address,
    scheduleMode: state.scheduledAt == null
        ? CheckoutScheduleMode.asap
        : CheckoutScheduleMode.scheduled,
    scheduledAt: state.scheduledAt,
    wantInvoice: state.wantInvoice,
    invoiceFields: state.invoiceFields,
    gift: state.gift,
    giftMessage: state.giftMessage ?? '',
    substitutionPreference: state.substitutionPreference,
    outOfStockRule: state.outOfStockRule,
    contactless: state.contactless,
    courierNote: state.courierNote ?? '',
    paymentType: state.paymentType,
    paymentMethodId: state.paymentMethodId,
    couponCode: state.couponCode,
    quote: state.quote,
  );
});

final checkoutListProvider =
    FutureProvider.autoDispose<List<CheckoutSession>>((ref) async {
  final result = await ref.watch(listCheckoutUseCaseProvider).call();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});

final checkoutQuoteProvider = Provider<CheckoutQuote?>((ref) {
  return ref.watch(checkoutControllerProvider.select((s) => s.quote));
});

final paymentMethodsListProvider =
    FutureProvider.autoDispose<List<SavedPaymentCard>>((ref) async {
  final result = await ref.watch(paymentMethodsRepositoryProvider).listSavedCards();
  return result.fold(
    onSuccess: (v) => v,
    onFailure: (e) => throw e,
  );
});
