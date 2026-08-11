import 'dart:async';
import 'dart:convert';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:uuid/uuid.dart';

import '../../../../di/providers.dart';
import '../../data/datasources/offers_remote_datasource.dart';
import '../../data/repositories/offers_repository_impl.dart';
import '../../domain/entities/offer_entity.dart';
import '../../domain/repositories/offers_repository.dart';

final offersRemoteDataSourceProvider = Provider<OffersRemoteDataSource>((ref) {
  return OffersRemoteDataSource(ref.watch(apiClientProvider));
});

final offersRepositoryProvider = Provider<OffersRepository>((ref) {
  return OffersRepositoryImpl(ref.watch(offersRemoteDataSourceProvider));
});

final offersProvider = FutureProvider.autoDispose<List<Offer>>((ref) async {
  final result = await ref.watch(offersRepositoryProvider).listOffers();
  return result.fold(onSuccess: (v) => v, onFailure: (e) => throw e);
});

/// Invalidates [offersProvider] when realtime events of type `offer.*` arrive.
final offersRealtimeInvalidationProvider = Provider.autoDispose<void>((ref) {
  final client = ref.watch(realtimeClientProvider);
  unawaited(client.connect());
  final sub = client.events.listen((event) {
    if (event is! RealtimeMessageEvent) return;
    if (_isOfferEvent(event.payload)) {
      ref.invalidate(offersProvider);
    }
  });
  ref.onDispose(sub.cancel);
});

bool _isOfferEvent(String payload) {
  try {
    final decoded = jsonDecode(payload);
    if (decoded is Map) {
      final type = decoded['type']?.toString() ??
          decoded['event']?.toString() ??
          decoded['event_type']?.toString() ??
          '';
      return type.startsWith('offer.');
    }
  } catch (_) {
    // Fall through to substring match for non-JSON payloads.
  }
  return payload.contains('offer.');
}

final offerActionsProvider = Provider((ref) => OfferActions(ref));

class OfferActions {
  OfferActions(this._ref);
  final Ref _ref;
  final _uuid = const Uuid();

  Future<Result<void>> accept(String offerId) {
    return _ref.read(offersRepositoryProvider).accept(
          offerId,
          idempotencyKey: _uuid.v4(),
        );
  }

  Future<Result<void>> reject(String offerId) {
    return _ref.read(offersRepositoryProvider).reject(
          offerId,
          idempotencyKey: _uuid.v4(),
        );
  }
}
