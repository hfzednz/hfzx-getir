import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../providers/categories_providers.dart';

class CategoryDetailScreen extends ConsumerWidget {
  const CategoryDetailScreen({super.key, required this.categoryId});

  final String categoryId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final products = ref.watch(categoryProductsProvider(categoryId));
    return Scaffold(
      appBar: NxTopBar(title: categoryId),
      body: products.when(
        data: (items) => GridView.builder(
          padding: const EdgeInsets.all(NxSpacing.s4),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 2,
            mainAxisSpacing: NxSpacing.s3,
            crossAxisSpacing: NxSpacing.s3,
            childAspectRatio: 0.72,
          ),
          itemCount: items.length,
          itemBuilder: (context, i) {
            final p = items[i];
            return NxProductCard(
              title: p.title,
              price: '${p.priceMinor / 100}',
              onTap: () => context.push('/p/${p.id}'),
            );
          },
        ),
        loading: () => const Center(child: NxSpinner()),
        error: (e, _) => Center(child: Text(e.toString())),
      ),
    );
  }
}
