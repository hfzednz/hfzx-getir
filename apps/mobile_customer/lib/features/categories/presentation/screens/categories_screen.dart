import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../../domain/entities/categories_entity.dart';
import '../providers/categories_providers.dart';
import '../../../../shared/errors/error_copy.dart';

class CategoriesScreen extends ConsumerWidget {
  const CategoriesScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(categoriesListProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.categoriesTitle),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(categoriesListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(categoriesListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                return _CategoryTile(item: item);
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: localizedCustomerError(context, e),
          onRetry: () => ref.invalidate(categoriesListProvider),
        ),
      ),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  const _CategoryTile({required this.item});

  final Category item;

  void _openDetail(BuildContext context) {
    context.push('/categories/${item.id}');
  }

  @override
  Widget build(BuildContext context) {
    final title = item.title.isNotEmpty ? item.title : item.id;
    final subtitle = item.productCount != null
        ? '${item.productCount} products'
        : null;
    final leading = _CategoryLeading(imageUrl: item.imageUrl ?? item.iconUrl);

    return NxCard(
      child: ListTile(
        leading: leading,
        title: Text(title),
        subtitle: subtitle != null ? Text(subtitle) : null,
        trailing: Icon(
          item.hasChildren ? Icons.folder_open_outlined : Icons.chevron_right,
        ),
        onTap: () => _openDetail(context),
      ),
    );
  }
}

class _CategoryLeading extends StatelessWidget {
  const _CategoryLeading({this.imageUrl});

  final String? imageUrl;

  @override
  Widget build(BuildContext context) {
    final colors = context.nxColors;
    if (imageUrl == null || imageUrl!.isEmpty) {
      return CircleAvatar(
        backgroundColor: colors.bgSunken,
        child: Icon(Icons.category_outlined, color: colors.iconSecondary),
      );
    }
    return ClipRRect(
      borderRadius: BorderRadius.circular(NxRadius.md),
      child: CachedNetworkImage(
        imageUrl: imageUrl!,
        width: 48,
        height: 48,
        fit: BoxFit.cover,
        placeholder: (_, __) => Container(
          width: 48,
          height: 48,
          color: colors.bgSunken,
          child: const Center(child: NxSpinner()),
        ),
        errorWidget: (_, __, ___) => CircleAvatar(
          backgroundColor: colors.bgSunken,
          child: Icon(Icons.broken_image_outlined, color: colors.iconSecondary),
        ),
      ),
    );
  }
}
