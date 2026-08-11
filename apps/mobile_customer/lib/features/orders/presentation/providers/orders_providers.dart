import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/orders_remote_datasource.dart';
import '../../data/repositories/orders_repository_impl.dart';
import '../../domain/entities/orders_entity.dart';
import '../../domain/repositories/orders_repository.dart';
import '../../domain/usecases/orders_usecases.dart';

final ordersRemoteDataSourceProvider = Provider<OrdersRemoteDataSource>((ref) {
  return OrdersRemoteDataSource(ref.watch(apiClientProvider));
});

final ordersRepositoryProvider = Provider<OrdersRepository>((ref) {
  return OrdersRepositoryImpl(ref.watch(ordersRemoteDataSourceProvider));
});

final getOrdersUseCaseProvider = Provider(
  (ref) => GetOrdersUseCase(ref.watch(ordersRepositoryProvider)),
);

final listOrdersUseCaseProvider = Provider(
  (ref) => ListOrdersUseCase(ref.watch(ordersRepositoryProvider)),
);

final cancelOrderUseCaseProvider = Provider(
  (ref) => CancelOrderUseCase(ref.watch(ordersRepositoryProvider)),
);

final partialCancelOrderUseCaseProvider = Provider(
  (ref) => PartialCancelOrderUseCase(ref.watch(ordersRepositoryProvider)),
);

final requestRefundUseCaseProvider = Provider(
  (ref) => RequestRefundUseCase(ref.watch(ordersRepositoryProvider)),
);

final reorderUseCaseProvider = Provider(
  (ref) => ReorderUseCase(ref.watch(ordersRepositoryProvider)),
);

final markFavoriteOrderUseCaseProvider = Provider(
  (ref) => MarkFavoriteOrderUseCase(ref.watch(ordersRepositoryProvider)),
);

final getOrderInvoiceUseCaseProvider = Provider(
  (ref) => GetOrderInvoiceUseCase(ref.watch(ordersRepositoryProvider)),
);

final getOrderReceiptUseCaseProvider = Provider(
  (ref) => GetOrderReceiptUseCase(ref.watch(ordersRepositoryProvider)),
);

final ordersListProvider = FutureProvider.autoDispose<List<Order>>((ref) async {
  final result = await ref.watch(listOrdersUseCaseProvider).call();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final orderDetailProvider =
    FutureProvider.autoDispose.family<Order, String>((ref, id) async {
  final result = await ref.watch(getOrdersUseCaseProvider).call(id: id);
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
