import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../../orders/domain/entities/orders_entity.dart';
import '../../../orders/presentation/providers/orders_providers.dart';
import '../providers/reviews_providers.dart';

class ReviewsScreen extends ConsumerStatefulWidget {
  const ReviewsScreen({super.key, this.orderId});

  final String? orderId;

  @override
  ConsumerState<ReviewsScreen> createState() => _ReviewsScreenState();
}

class _ReviewsScreenState extends ConsumerState<ReviewsScreen> {
  int _orderRating = 5;
  int _courierRating = 5;
  final _productRatings = <String, int>{};
  final _commentController = TextEditingController();
  final _photos = <XFile>[];
  final _picker = ImagePicker();
  bool _seededProductRatings = false;

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  Future<void> _pickPhotos() async {
    final files = await _picker.pickMultiImage(imageQuality: 80);
    if (files.isNotEmpty) setState(() => _photos.addAll(files));
  }

  void _ensureProductRatings(List<OrderLineItem> items) {
    if (_seededProductRatings || items.isEmpty) return;
    for (final item in items) {
      final key = item.productId.isNotEmpty ? item.productId : item.id;
      _productRatings.putIfAbsent(key, () => 5);
    }
    _seededProductRatings = true;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final orderId = widget.orderId;
    final submitState = ref.watch(reviewSubmitProvider);

    if (orderId == null) {
      return Scaffold(
        appBar: NxTopBar(title: l10n.reviewsTitle),
        body: const Center(child: Text('Open from an order to submit a review')),
      );
    }

    final orderAsync = ref.watch(orderDetailProvider(orderId));
    final orderItems = orderAsync.maybeWhen(
      data: (order) => order.items,
      orElse: () => const <OrderLineItem>[],
    );
    _ensureProductRatings(orderItems);

    return Scaffold(
      appBar: NxTopBar(title: l10n.reviewsTitle),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: [
          Text('Rate your order', style: NxTypography.headlineSm),
          _StarRow(rating: _orderRating, onChanged: (v) => setState(() => _orderRating = v)),
          const SizedBox(height: NxSpacing.s4),
          Text('Rate courier', style: NxTypography.headlineSm),
          _StarRow(rating: _courierRating, onChanged: (v) => setState(() => _courierRating = v)),
          if (orderItems.isNotEmpty) ...[
            const SizedBox(height: NxSpacing.s4),
            Text('Rate products', style: NxTypography.headlineSm),
            const SizedBox(height: NxSpacing.s2),
            ...orderItems.map((item) {
              final key = item.productId.isNotEmpty ? item.productId : item.id;
              final rating = _productRatings[key] ?? 5;
              return Padding(
                padding: const EdgeInsets.only(bottom: NxSpacing.s3),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(item.name, style: NxTypography.bodyMd),
                    _StarRow(
                      rating: rating,
                      onChanged: (v) => setState(() => _productRatings[key] = v),
                    ),
                  ],
                ),
              );
            }),
          ],
          const SizedBox(height: NxSpacing.s4),
          TextField(
            controller: _commentController,
            decoration: const InputDecoration(labelText: 'Comment', border: OutlineInputBorder()),
            maxLines: 4,
          ),
          const SizedBox(height: NxSpacing.s3),
          Wrap(
            spacing: NxSpacing.s2,
            children: [
              ..._photos.map((p) => Image.file(File(p.path), width: 72, height: 72, fit: BoxFit.cover)),
              OutlinedButton.icon(onPressed: _pickPhotos, icon: const Icon(Icons.photo), label: const Text('Add photos')),
            ],
          ),
          const SizedBox(height: NxSpacing.s4),
          NxBanner(
            title: 'Verified purchase',
            message: 'This review will be marked as verified for order $orderId',
            variant: NxBannerVariant.info,
          ),
          const SizedBox(height: NxSpacing.s4),
          if (submitState.hasError)
            ErrorView(message: submitState.error.toString(), onRetry: () => ref.invalidate(reviewSubmitProvider)),
          NxButton(
            label: 'Submit review',
            loading: submitState.isLoading,
            onPressed: () => ref.read(reviewSubmitProvider.notifier).submit(
                  orderId: orderId,
                  orderRating: _orderRating,
                  courierRating: _courierRating,
                  productRatings: Map<String, int>.from(_productRatings),
                  comment: _commentController.text,
                  photos: _photos,
                ),
          ),
        ],
      ),
    );
  }
}

class _StarRow extends StatelessWidget {
  const _StarRow({required this.rating, required this.onChanged});

  final int rating;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: List.generate(5, (i) {
        final star = i + 1;
        return IconButton(
          icon: Icon(star <= rating ? Icons.star : Icons.star_border, color: context.nxColors.warning),
          onPressed: () => onChanged(star),
        );
      }),
    );
  }
}
