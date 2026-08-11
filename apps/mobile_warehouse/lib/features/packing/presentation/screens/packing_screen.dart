import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'pack_task_screen.dart';
import 'packing_queue_screen.dart';

/// Shell entry for packing — production queue UI.
class PackingScreen extends ConsumerWidget {
  const PackingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return const PackingQueueScreen();
  }
}

class PackingTaskScreen extends ConsumerWidget {
  const PackingTaskScreen({super.key, required this.taskId});

  final String taskId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PackTaskScreen(taskId: taskId);
  }
}
