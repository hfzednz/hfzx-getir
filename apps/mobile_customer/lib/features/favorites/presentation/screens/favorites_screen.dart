import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/favorites_entity.dart';
import '../providers/favorites_providers.dart';

class FavoritesScreen extends ConsumerWidget {
  const FavoritesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final filter = ref.watch(favoriteTypeFilterProvider);
    final entriesAsync = ref.watch(favoriteEntriesStreamProvider(filter));

    ref.listen(favoritesSyncProvider, (_, __) {});

    return Scaffold(
      appBar: NxTopBar(title: l10n.favoritesTitle),
      body: Column(
        children: [
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.all(NxSpacing.s4),
            child: Row(
              children: [
                _FilterChip(label: 'All', selected: filter == null, onTap: () => ref.read(favoriteTypeFilterProvider.notifier).state = null),
                ...FavoriteType.values.map(
                  (t) => _FilterChip(
                    label: t.name,
                    selected: filter == t,
                    onTap: () => ref.read(favoriteTypeFilterProvider.notifier).state = t,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: AsyncValueWidget(
              value: entriesAsync,
              data: (items) {
                if (items.isEmpty) {
                  return NxEmptyState(
                    title: l10n.emptyTitle,
                    body: l10n.emptyMessage,
                    primaryActionLabel: l10n.retry,
                    onPrimaryAction: () => ref.invalidate(favoritesSyncProvider),
                  );
                }
                return RefreshIndicator(
                  onRefresh: () async => ref.invalidate(favoritesSyncProvider),
                  child: ListView.separated(
                    padding: const EdgeInsets.all(NxSpacing.s4),
                    itemCount: items.length,
                    separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
                    itemBuilder: (context, index) {
                      final item = items[index];
                      return NxCard(
                        child: ListTile(
                          leading: item.imageUrl != null
                              ? Image.network(item.imageUrl!, width: 48, height: 48, fit: BoxFit.cover)
                              : Icon(_iconForType(item.type)),
                          title: Text(item.title.isNotEmpty ? item.title : item.targetId),
                          subtitle: Text('${item.type.name}${item.subtitle.isNotEmpty ? ' · ${item.subtitle}' : ''}'),
                          trailing: IconButton(
                            icon: const Icon(Icons.delete_outline),
                            onPressed: () => ref.read(favoritesRepositoryProvider).remove(item.id),
                          ),
                        ),
                      );
                    },
                  ),
                );
              },
              error: (e, _) => Center(child: Text(e.toString())),
            ),
          ),
        ],
      ),
    );
  }

  IconData _iconForType(FavoriteType type) => switch (type) {
        FavoriteType.product => Icons.shopping_bag_outlined,
        FavoriteType.brand => Icons.storefront_outlined,
        FavoriteType.category => Icons.category_outlined,
        FavoriteType.search => Icons.search,
        FavoriteType.store => Icons.store_outlined,
      };
}

class _FilterChip extends StatelessWidget {
  const _FilterChip({required this.label, required this.selected, required this.onTap});

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(right: NxSpacing.s2),
      child: FilterChip(label: Text(label), selected: selected, onSelected: (_) => onTap()),
    );
  }
}
