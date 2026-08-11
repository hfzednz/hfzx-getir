export 'pickup_scan_screen.dart' show PickupScanScreen;

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'pickup_scan_screen.dart';

/// Router-facing alias for [PickupScanScreen].
class DeliveryPickupScreen extends ConsumerWidget {
  const DeliveryPickupScreen({super.key, required this.deliveryId});

  final String deliveryId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return PickupScanScreen(deliveryId: deliveryId);
  }
}
