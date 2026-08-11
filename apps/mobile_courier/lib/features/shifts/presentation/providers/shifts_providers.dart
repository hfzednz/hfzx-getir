import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/shifts_remote_datasource.dart';
import '../../data/repositories/shifts_repository_impl.dart';
import '../../domain/entities/shift_entity.dart';
import '../../domain/repositories/shifts_repository.dart';

final shiftsRemoteDataSourceProvider = Provider<ShiftsRemoteDataSource>((ref) {
  return ShiftsRemoteDataSource(ref.watch(apiClientProvider));
});

final shiftsRepositoryProvider = Provider<ShiftsRepository>((ref) {
  return ShiftsRepositoryImpl(ref.watch(shiftsRemoteDataSourceProvider));
});

final shiftsProvider =
    FutureProvider.autoDispose<List<CourierShift>>((ref) async {
  final result = await ref.watch(shiftsRepositoryProvider).listShifts();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

final shiftActionsProvider = Provider((ref) => ShiftActions(ref));

class ShiftActions {
  ShiftActions(this._ref);
  final Ref _ref;

  Future<void> start() async {
    await _ref.read(shiftsRepositoryProvider).startShift();
    _ref.invalidate(shiftsProvider);
  }

  Future<void> end(String id) async {
    await _ref.read(shiftsRepositoryProvider).endShift(id);
    _ref.invalidate(shiftsProvider);
  }

  Future<void> startBreak(String id) async {
    await _ref.read(shiftsRepositoryProvider).startBreak(id);
    _ref.invalidate(shiftsProvider);
  }

  Future<void> endBreak(String id) async {
    await _ref.read(shiftsRepositoryProvider).endBreak(id);
    _ref.invalidate(shiftsProvider);
  }
}
