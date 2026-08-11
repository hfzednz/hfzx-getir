import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../../auth/presentation/providers/auth_session_provider.dart';
import '../../data/datasources/shifts_remote_datasource.dart';
import '../../data/repositories/shifts_repository_impl.dart';
import '../../domain/entities/shift_entity.dart';
import '../../domain/repositories/shifts_repository.dart';

final shiftsRemoteDataSourceProvider =
    Provider((ref) => ShiftsRemoteDataSource(ref.watch(apiClientProvider)));
final shiftsRepositoryProvider = Provider<ShiftsRepository>(
  (ref) => ShiftsRepositoryImpl(ref.watch(shiftsRemoteDataSourceProvider)),
);
final currentShiftProvider =
    FutureProvider.autoDispose<WarehouseShift?>((ref) async {
  final r = await ref.watch(shiftsRepositoryProvider).current();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final attendanceProvider =
    FutureProvider.autoDispose<List<WarehouseShift>>((ref) async {
  final r = await ref.watch(shiftsRepositoryProvider).attendance();
  return r.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});
final shiftsActionsProvider = Provider((ref) => ShiftsActions(ref));

class ShiftsActions {
  ShiftsActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  void _inv() {
    _ref.invalidate(currentShiftProvider);
    _ref.invalidate(attendanceProvider);
  }

  Future<Result<WarehouseShift>> clockIn() async {
    final r = await _ref
        .read(shiftsRepositoryProvider)
        .clockIn(idempotencyKey: _uuid.v4());
    if (r.isSuccess) {
      _inv();
      final id = r.valueOrNull?.id;
      if (id != null) {
        await _ref.read(authSessionProvider.notifier).updateShift(shiftId: id);
      }
    }
    return r;
  }

  Future<Result<WarehouseShift>> clockOut() async {
    final r = await _ref
        .read(shiftsRepositoryProvider)
        .clockOut(idempotencyKey: _uuid.v4());
    if (r.isSuccess) {
      _inv();
      await _ref.read(authSessionProvider.notifier).clearShift();
    }
    return r;
  }

  Future<Result<WarehouseShift>> startBreak() async {
    final r = await _ref
        .read(shiftsRepositoryProvider)
        .startBreak(idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv();
    return r;
  }

  Future<Result<WarehouseShift>> endBreak() async {
    final r = await _ref
        .read(shiftsRepositoryProvider)
        .endBreak(idempotencyKey: _uuid.v4());
    if (r.isSuccess) _inv();
    return r;
  }
}
