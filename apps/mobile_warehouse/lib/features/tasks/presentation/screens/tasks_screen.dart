import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';
import '../../../../shared/widgets/async_value_widget.dart';
import '../../domain/entities/ops_task.dart';
import '../providers/tasks_providers.dart';

class TasksScreen extends ConsumerWidget {
  const TasksScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final async = ref.watch(tasksListProvider);
    return Scaffold(
      appBar: const NxTopBar(title: 'Tasks'),
      body: RefreshIndicator(
        onRefresh: () async => ref.invalidate(tasksListProvider),
        child: AsyncValueWidget<List<OpsTask>>(
          value: async,
          data: (items) {
            if (items.isEmpty) {
              return ListView(children: const [SizedBox(height: 120), NxEmptyState(title: 'No tasks')]);
            }
            return ListView.separated(
              padding: const EdgeInsets.all(NxSpacing.s3),
              itemCount: items.length,
              separatorBuilder: (_, __) => const SizedBox(height: NxSpacing.s2),
              itemBuilder: (context, i) {
                final t = items[i];
                return NxCard(
                  child: Row(
                    children: [
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(t.title, style: NxTypography.titleSm),
                            Text('${t.category} · ${t.status} · p${t.priority}', style: NxTypography.captionMd),
                          ],
                        ),
                      ),
                      if (t.status != 'done')
                        NxButton(label: 'Done', size: NxButtonSize.sm, onPressed: () => ref.read(tasksActionsProvider).complete(t.id)),
                    ],
                  ),
                );
              },
            );
          },
        ),
      ),
    );
  }
}
