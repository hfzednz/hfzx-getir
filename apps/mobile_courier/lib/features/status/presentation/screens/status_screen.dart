import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_design/nexora_design.dart';

import '../providers/duty_controller.dart';
import '../widgets/status_control_widget.dart';

class StatusScreen extends ConsumerStatefulWidget {
  const StatusScreen({super.key});

  @override
  ConsumerState<StatusScreen> createState() => _StatusScreenState();
}

class _StatusScreenState extends ConsumerState<StatusScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() => ref.read(dutyControllerProvider.notifier).refresh());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: const NxTopBar(title: 'Duty status'),
      body: ListView(
        padding: const EdgeInsets.all(NxSpacing.s4),
        children: const [
          StatusControlWidget(),
        ],
      ),
    );
  }
}
