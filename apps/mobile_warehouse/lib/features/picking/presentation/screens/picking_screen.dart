import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'pick_scan_screen.dart';
import 'pick_task_screen.dart';
import 'picking_queue_screen.dart';

/// Shell entry for picking — production queue UI.
class PickingScreen extends ConsumerWidget {
  const PickingScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return const PickingQueueScreen();
  }
}

class PickingTaskScreen extends ConsumerWidget {
  const PickingTaskScreen({super.key, required this.taskId});

  final String taskId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PickTaskScreen(taskId: taskId);
  }
}

class PickingScanScreen extends ConsumerWidget {
  const PickingScanScreen({super.key, required this.taskId});

  final String taskId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PickScanScreen(taskId: taskId);
  }
}
