import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/entities/warehouse_session.dart';
import 'auth_session_provider.dart';

/// Public alias for warehouse session snapshot (ARCHITECTURE.md).
final warehouseSessionProvider = Provider<WarehouseSession?>((ref) {
  final session = ref.watch(authSessionProvider);
  if (!session.isAuthenticated || session.userId == null) return null;
  return WarehouseSession(
    userId: session.userId!,
    accessToken: '',
    refreshToken: '',
    role: session.role,
    storeId: session.storeId ?? '',
    stationId: session.stationId,
    shiftId: session.shiftId,
    displayName: session.displayName,
    phone: session.phone,
    kycOk: session.kycOk,
    deviceOk: session.deviceOk,
  );
});
