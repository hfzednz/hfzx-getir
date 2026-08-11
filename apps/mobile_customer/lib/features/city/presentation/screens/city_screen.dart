import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../../../../di/providers.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../../../shared/widgets/error_view.dart';
import '../providers/city_providers.dart';

class CityScreen extends ConsumerWidget {
  const CityScreen({super.key, this.id});

  final String? id;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final asyncItems = ref.watch(cityListProvider);
    final selectedId = ref.watch(cityIdProvider);

    return Scaffold(
      appBar: NxTopBar(title: l10n.cityTitle),
      body: AsyncValueWidget(
        value: asyncItems,
        data: (items) {
          if (items.isEmpty) {
            return NxEmptyState(
              title: l10n.emptyTitle,
              body: l10n.emptyMessage,
              primaryActionLabel: l10n.retry,
              onPrimaryAction: () => ref.invalidate(cityListProvider),
            );
          }
          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(cityListProvider),
            child: ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s4),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s3),
              itemBuilder: (context, index) {
                final item = items[index];
                final selected = item.id == selectedId;
                return NxCard(
                  child: ListTile(
                    title: Text(item.title.isNotEmpty ? item.title : item.id),
                    subtitle: item.title.isNotEmpty ? Text(item.id) : null,
                    trailing: selected
                        ? Icon(Icons.check_circle, color: context.nxColors.textBrand)
                        : const Icon(Icons.chevron_right),
                    onTap: () {
                      ref.read(cityIdProvider.notifier).state = item.id;
                      ref.read(preferencesStoreProvider).set('city_id', item.id);
                      Navigator.of(context).pop(item.id);
                    },
                  ),
                );
              },
            ),
          );
        },
        error: (e, _) => ErrorView(
          message: e.toString(),
          onRetry: () => ref.invalidate(cityListProvider),
        ),
      ),
    );
  }
}
