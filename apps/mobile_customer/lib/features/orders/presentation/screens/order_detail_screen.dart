import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/business_rules/order_rules.dart';
import '../../../../shared/utils/idempotency.dart';
import '../../domain/entities/orders_entity.dart';
import '../providers/orders_providers.dart';
import '../../../../di/analytics_providers.dart';
import '../../../../shared/analytics/analytics_events.dart';
import '../../../cart/presentation/providers/cart_providers.dart';

class OrderDetailScreen extends ConsumerStatefulWidget {
  const OrderDetailScreen({super.key, required this.orderId});

  final String orderId;

  @override
  ConsumerState<OrderDetailScreen> createState() => _OrderDetailScreenState();
}

class _OrderDetailScreenState extends ConsumerState<OrderDetailScreen> {
  bool _busy = false;

  Future<void> _runAction(Future<void> Function() action) async {
    setState(() => _busy = true);
    try {
      await action();
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final orderAsync = ref.watch(orderDetailProvider(widget.orderId));

    return Scaffold(
      appBar: NxTopBar(title: l10n.ordersTitle),
      body: orderAsync.when(
        data: (order) => Stack(
          children: [
            SingleChildScrollView(
              padding: const EdgeInsets.all(NxSpacing.s4),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  NxOrderCard(
                    orderId: order.id,
                    statusLabel: order.statusLabel,
                    total: order.totalLabel,
                    subtitle: order.etaMinutes != null
                        ? '${l10n.deliveryEta} ~ ${order.etaMinutes} min'
                        : null,
                    imageUrls: order.items
                        .map((i) => i.imageUrl ?? '')
                        .where((u) => u.isNotEmpty)
                        .toList(),
                    extraItemCount: order.items.length > 4 ? order.items.length - 4 : 0,
                  ),
                  if (order.courier?.name != null) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxCard(
                      child: ListTile(
                        leading: const Icon(Icons.delivery_dining),
                        title: Text(order.courier!.name ?? l10n.courierLabel),
                        subtitle: Text(order.courier!.vehicle ?? ''),
                      ),
                    ),
                  ],
                  if (order.timeline.isNotEmpty) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxCard(
                      child: Padding(
                        padding: const EdgeInsets.all(NxSpacing.s3),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Timeline', style: NxTypography.titleSm),
                            const SizedBox(height: NxSpacing.s2),
                            ...order.timeline.map(
                              (e) => ListTile(
                                dense: true,
                                contentPadding: EdgeInsets.zero,
                                title: Text(e.label),
                                subtitle: e.subtitle != null ? Text(e.subtitle!) : null,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                  if (order.status == OrderLifecycleStatus.delivered &&
                      order.hasProofOfDelivery) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxCard(
                      child: Padding(
                        padding: const EdgeInsets.all(NxSpacing.s3),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(l10n.proofOfDelivery, style: NxTypography.titleSm),
                            const SizedBox(height: NxSpacing.s2),
                            Wrap(
                              spacing: NxSpacing.s2,
                              runSpacing: NxSpacing.s2,
                              children: [
                                ...order.proofOfDeliveryPhotos.map(
                                  (url) => GestureDetector(
                                    onTap: () => launchUrl(Uri.parse(url)),
                                    child: ClipRRect(
                                      borderRadius: BorderRadius.circular(8),
                                      child: Image.network(
                                        url,
                                        width: 88,
                                        height: 88,
                                        fit: BoxFit.cover,
                                        errorBuilder: (_, __, ___) => const SizedBox(
                                          width: 88,
                                          height: 88,
                                          child: ColoredBox(
                                            color: Color(0x11000000),
                                            child: Icon(Icons.broken_image),
                                          ),
                                        ),
                                      ),
                                    ),
                                  ),
                                ),
                                if (order.proofOfDeliveryUrl != null &&
                                    order.proofOfDeliveryUrl!.isNotEmpty &&
                                    !order.proofOfDeliveryPhotos
                                        .contains(order.proofOfDeliveryUrl))
                                  GestureDetector(
                                    onTap: () => launchUrl(
                                      Uri.parse(order.proofOfDeliveryUrl!),
                                    ),
                                    child: ClipRRect(
                                      borderRadius: BorderRadius.circular(8),
                                      child: Image.network(
                                        order.proofOfDeliveryUrl!,
                                        width: 88,
                                        height: 88,
                                        fit: BoxFit.cover,
                                        errorBuilder: (_, __, ___) => const SizedBox(
                                          width: 88,
                                          height: 88,
                                          child: ColoredBox(
                                            color: Color(0x11000000),
                                            child: Icon(Icons.image),
                                          ),
                                        ),
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: NxSpacing.s4),
                  NxButton(
                    label: l10n.trackOrder,
                    expand: true,
                    onPressed: () => context.push('/orders/${widget.orderId}/track'),
                  ),
                  const SizedBox(height: NxSpacing.s3),
                  if (order.canCancel)
                    NxButton(
                      label: l10n.cancelOrder,
                      variant: NxButtonVariant.secondary,
                      expand: true,
                      onPressed: _busy
                          ? null
                          : () => _runAction(() => _cancelOrder(order)),
                    )
                  else
                    Padding(
                      padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                      child: Text(
                        order.cancellationPolicy.policyText ??
                            l10n.cannotCancelOrder,
                        style: NxTypography.captionMd,
                      ),
                    ),
                  if (order.canPartialCancel) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: l10n.partialCancel,
                      variant: NxButtonVariant.secondary,
                      expand: true,
                      onPressed: _busy
                          ? null
                          : () => _runAction(() => _partialCancel(order)),
                    ),
                  ],
                  if (order.cancellationPolicy.refundEligible) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: l10n.requestRefund,
                      variant: NxButtonVariant.secondary,
                      expand: true,
                      onPressed: _busy
                          ? null
                          : () => _runAction(() => _requestRefund(order)),
                    ),
                  ],
                  if (order.canReorder) ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: l10n.reorder,
                      expand: true,
                      onPressed: _busy ? null : () => _runAction(() => _reorder(order)),
                    ),
                  ],
                  const SizedBox(height: NxSpacing.s3),
                  Row(
                    children: [
                      Expanded(
                        child: NxButton(
                          label: order.isFavorite ? l10n.unfavorite : l10n.favorite,
                          variant: NxButtonVariant.tertiary,
                          onPressed: _busy
                              ? null
                              : () => _runAction(
                                    () => _toggleFavorite(order),
                                  ),
                        ),
                      ),
                      const SizedBox(width: NxSpacing.s3),
                      Expanded(
                        child: NxButton(
                          label: l10n.invoice,
                          variant: NxButtonVariant.tertiary,
                          onPressed: order.invoiceUrl != null
                              ? () => launchUrl(Uri.parse(order.invoiceUrl!))
                              : () => _runAction(() => _openInvoice(order.id)),
                        ),
                      ),
                    ],
                  ),
                  if (order.receiptUrl != null || order.status.name == 'delivered') ...[
                    const SizedBox(height: NxSpacing.s3),
                    NxButton(
                      label: l10n.receipt,
                      variant: NxButtonVariant.tertiary,
                      expand: true,
                      onPressed: order.receiptUrl != null
                          ? () => launchUrl(Uri.parse(order.receiptUrl!))
                          : () => _runAction(() => _openReceipt(order.id)),
                    ),
                  ],
                ],
              ),
            ),
            if (_busy)
              const Positioned.fill(
                child: ColoredBox(
                  color: Color(0x33000000),
                  child: Center(child: NxSpinner()),
                ),
              ),
          ],
        ),
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }

  Future<void> _cancelOrder(Order order) async {
    final result = await ref.read(cancelOrderUseCaseProvider).call(
          order: order,
          idempotencyKey: Idempotency.generate(),
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (_) {
        // ignore: unawaited_futures
        ref.read(analyticsTrackerProvider).trackRaw(
              eventName: AnalyticsEvents.orderCancelled,
              props: {'order_id': order.id},
            );
        ref.invalidate(orderDetailProvider(widget.orderId));
        NxToast.show(context, message: 'Order cancelled');
      },
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _partialCancel(Order order) async {
    final activeLines = order.items
        .where((i) => !i.cancelled)
        .map((i) => i.id)
        .toList();
    if (activeLines.isEmpty) return;

    final result = await ref.read(partialCancelOrderUseCaseProvider).call(
          order: order,
          lineIds: [activeLines.first],
          idempotencyKey: Idempotency.generate(),
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (_) {
        ref.invalidate(orderDetailProvider(widget.orderId));
        NxToast.show(context, message: 'Items cancelled');
      },
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _requestRefund(Order order) async {
    final result = await ref.read(requestRefundUseCaseProvider).call(
          id: order.id,
          idempotencyKey: Idempotency.generate(),
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (_) => NxToast.show(context, message: 'Refund requested'),
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _reorder(Order order) async {
    final result = await ref.read(reorderUseCaseProvider).call(
          order: order,
          idempotencyKey: Idempotency.generate(),
        );
    if (!mounted) return;
    await result.fold(
      onSuccess: (seeded) async {
        final lines = seeded.items.isNotEmpty ? seeded.items : order.items;
        var added = 0;
        var skipped = 0;
        for (final line in lines) {
          if (line.cancelled || line.productId.isEmpty) {
            skipped++;
            continue;
          }
          await ref.read(cartRepositoryProvider).addItem(
                productId: line.productId,
                title: line.name.isNotEmpty ? line.name : line.productId,
                imageUrl: line.imageUrl,
                unitPriceMinor: line.unitPriceMinor,
                quantity: line.quantity,
              );
          added++;
        }
        if (!mounted) return;
        if (added == 0) {
          NxToast.show(
            context,
            message: skipped > 0
                ? 'Those items are no longer available'
                : 'Could not add items to cart',
          );
          return;
        }
        NxToast.show(context, message: 'Items added to cart');
        context.push('/cart');
      },
      onFailure: (e) async => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _toggleFavorite(Order order) async {
    final result = await ref.read(markFavoriteOrderUseCaseProvider).call(
          order.id,
          favorite: !order.isFavorite,
        );
    if (!mounted) return;
    result.fold(
      onSuccess: (_) => ref.invalidate(orderDetailProvider(widget.orderId)),
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _openInvoice(String id) async {
    final result = await ref.read(getOrderInvoiceUseCaseProvider).call(id);
    if (!mounted) return;
    result.fold(
      onSuccess: (doc) => launchUrl(Uri.parse(doc.url)),
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }

  Future<void> _openReceipt(String id) async {
    final result = await ref.read(getOrderReceiptUseCaseProvider).call(id);
    if (!mounted) return;
    result.fold(
      onSuccess: (doc) => launchUrl(Uri.parse(doc.url)),
      onFailure: (e) => NxToast.show(context, message: e.message),
    );
  }
}
