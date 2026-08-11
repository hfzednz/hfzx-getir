import 'package:equatable/equatable.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';

import '../../../../di/providers.dart';
import '../../../../shared/business_rules/duty_rules.dart';
import '../../data/datasources/duty_remote_datasource.dart';
import '../../data/repositories/duty_repository_impl.dart';
import '../../domain/entities/duty_status.dart';
import '../../domain/repositories/duty_repository.dart';

final dutyRemoteDataSourceProvider = Provider<DutyRemoteDataSource>((ref) {
  return DutyRemoteDataSource(ref.watch(apiClientProvider));
});

final dutyRepositoryProvider = Provider<DutyRepository>((ref) {
  return DutyRepositoryImpl(ref.watch(dutyRemoteDataSourceProvider));
});

class DutyState extends Equatable {
  const DutyState({
    this.status = DutyStatus.offline,
    this.isLoading = false,
    this.error,
    this.hasActiveDelivery = false,
  });

  final DutyStatus status;
  final bool isLoading;
  final String? error;
  final bool hasActiveDelivery;

  DutyState copyWith({
    DutyStatus? status,
    bool? isLoading,
    String? error,
    bool? hasActiveDelivery,
  }) {
    return DutyState(
      status: status ?? this.status,
      isLoading: isLoading ?? this.isLoading,
      error: error,
      hasActiveDelivery: hasActiveDelivery ?? this.hasActiveDelivery,
    );
  }

  @override
  List<Object?> get props => [status, isLoading, error, hasActiveDelivery];
}

class DutyController extends StateNotifier<DutyState> {
  DutyController(this._repository) : super(const DutyState());

  final DutyRepository _repository;

  Future<void> refresh() async {
    state = state.copyWith(isLoading: true);
    final result = await _repository.getStatus();
    result.fold(
      onSuccess: (status) {
        state = state.copyWith(status: status, isLoading: false);
      },
      onFailure: (error) {
        state = state.copyWith(isLoading: false, error: error.message);
      },
    );
  }

  void setHasActiveDelivery(bool value) {
    state = state.copyWith(hasActiveDelivery: value);
  }

  Future<Result<DutyStatus>> setStatus(DutyStatus next) async {
    final validation = DutyRules.validateTransition(
      from: state.status,
      to: next,
      hasActiveDelivery: state.hasActiveDelivery,
    );
    if (validation.isFailure) {
      state = state.copyWith(error: validation.errorOrNull?.message);
      return Failure(validation.errorOrNull!);
    }

    state = state.copyWith(isLoading: true);
    final result = await _repository.setStatus(next);
    return result.fold(
      onSuccess: (status) {
        state = state.copyWith(status: status, isLoading: false);
        return Success(status);
      },
      onFailure: (error) {
        state = state.copyWith(isLoading: false, error: error.message);
        return Failure(error);
      },
    );
  }

  Future<Result<DutyStatus>> toggleOnline() {
    final next = state.status == DutyStatus.offline
        ? DutyStatus.online
        : DutyStatus.offline;
    return setStatus(next);
  }
}

final dutyControllerProvider =
    StateNotifierProvider<DutyController, DutyState>((ref) {
  return DutyController(ref.watch(dutyRepositoryProvider));
});
