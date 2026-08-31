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
import '../../../../shared/errors/error_copy.dart';

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
        body: Center(child: Text(l10n.openReviewFromOrder)),
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
          Text(l10n.rateYourOrder, style: NxTypography.headlineSm),
          _StarRow(rating: _orderRating, onChanged: (v) => setState(() => _orderRating = v)),
          const SizedBox(height: NxSpacing.s4),
          Text(l10n.rateCourier, style: NxTypography.headlineSm),
          _StarRow(rating: _courierRating, onChanged: (v) => setState(() => _courierRating = v)),
          if (orderItems.isNotEmpty) ...[
            const SizedBox(height: NxSpacing.s4),
            Text(l10n.rateProducts, style: NxTypography.headlineSm),
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
            decoration: InputDecoration(labelText: l10n.comment, border: const OutlineInputBorder()),
            maxLines: 4,
          ),
          const SizedBox(height: NxSpacing.s3),
          Wrap(
            spacing: NxSpacing.s2,
            children: [
              ..._photos.map((p) => Image.file(File(p.path), width: 72, height: 72, fit: BoxFit.cover)),
              OutlinedButton.icon(onPressed: _pickPhotos, icon: const Icon(Icons.photo), label: Text(l10n.addPhotos)),
            ],
          ),
          const SizedBox(height: NxSpacing.s4),
          NxBanner(
            title: l10n.verifiedPurchase,
            message: '${l10n.verifiedPurchase} $orderId',
            variant: NxBannerVariant.info,
          ),
          const SizedBox(height: NxSpacing.s4),
          if (submitState.hasError)
            ErrorView(message: localizedCustomerError(context, submitState.error!), onRetry: () => ref.invalidate(reviewSubmitProvider)),
          NxButton(
            label: l10n.submitReview,
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
