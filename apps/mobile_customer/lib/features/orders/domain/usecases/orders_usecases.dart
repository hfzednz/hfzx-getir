import 'package:nexora_core/nexora_core.dart';

import '../../../../shared/business_rules/order_rules.dart';
import '../entities/orders_entity.dart';
import '../repositories/orders_repository.dart';

class GetOrdersUseCase {
  const GetOrdersUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call({String? id}) => _repository.fetch(id: id);
}

class ListOrdersUseCase {
  const ListOrdersUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<List<Order>>> call({Map<String, dynamic>? params}) =>
      _repository.fetchList(params: params);
}

class CancelOrderUseCase {
  const CancelOrderUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call({
    required Order order,
    String? reason,
    String? idempotencyKey,
  }) {
    final validation = OrderRules.validateCancel(
      status: order.status,
      paymentCaptured: order.paymentCaptured,
    );
    if (validation.isFailure) {
      return Future.value(Failure(validation.errorOrNull!));
    }
    return _repository.cancelOrder(
      id: order.id,
      reason: reason,
      idempotencyKey: idempotencyKey,
    );
  }
}

class PartialCancelOrderUseCase {
  const PartialCancelOrderUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call({
    required Order order,
    required List<String> lineIds,
    String? idempotencyKey,
  }) {
    final validation = OrderRules.validatePartialCancel(
      status: order.status,
      lineIdsToCancel: lineIds,
      totalLineCount: order.items.length,
    );
    if (validation.isFailure) {
      return Future.value(Failure(validation.errorOrNull!));
    }
    return _repository.partialCancel(
      id: order.id,
      lineIds: lineIds,
      idempotencyKey: idempotencyKey,
    );
  }
}

class RequestRefundUseCase {
  const RequestRefundUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call({
    required String id,
    String? reason,
    String? idempotencyKey,
  }) =>
      _repository.requestRefund(id: id, reason: reason, idempotencyKey: idempotencyKey);
}

class ReorderUseCase {
  const ReorderUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call({
    required Order order,
    String? idempotencyKey,
  }) {
    final validation = OrderRules.validateReorder(
      status: order.status,
      allItemsAvailable: order.allItemsAvailable,
    );
    if (validation.isFailure) {
      return Future.value(Failure(validation.errorOrNull!));
    }
    return _repository.reorder(id: order.id, idempotencyKey: idempotencyKey);
  }
}

class MarkFavoriteOrderUseCase {
  const MarkFavoriteOrderUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<Order>> call(String id, {required bool favorite}) =>
      _repository.markFavoriteOrder(id, favorite: favorite);
}

class GetOrderInvoiceUseCase {
  const GetOrderInvoiceUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<OrderDocument>> call(String id) => _repository.getInvoice(id);
}

class GetOrderReceiptUseCase {
  const GetOrderReceiptUseCase(this._repository);
  final OrdersRepository _repository;

  Future<Result<OrderDocument>> call(String id) => _repository.getReceipt(id);
}
