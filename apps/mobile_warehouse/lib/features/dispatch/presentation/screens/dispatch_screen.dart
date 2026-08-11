import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'dispatch_queue_screen.dart';
import 'handoff_detail_screen.dart';

/// Shell entry for dispatch — production queue UI.
class DispatchScreen extends ConsumerWidget {
  const DispatchScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return const DispatchQueueScreen();
  }
}

class DispatchHandoffScreen extends ConsumerWidget {
  const DispatchHandoffScreen({super.key, required this.handoffId});

  final String handoffId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return HandoffDetailScreen(handoffId: handoffId);
  }
}
