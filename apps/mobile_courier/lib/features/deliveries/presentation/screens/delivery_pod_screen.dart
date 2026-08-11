export 'pod_screen.dart' show PodScreen;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'pod_screen.dart';

/// Router-facing alias for [PodScreen].
class DeliveryPodScreen extends ConsumerWidget {
  const DeliveryPodScreen({super.key, required this.deliveryId});

  final String deliveryId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PodScreen(deliveryId: deliveryId);
  }
}
