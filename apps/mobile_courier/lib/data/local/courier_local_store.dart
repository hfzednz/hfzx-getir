import 'dart:convert';

import 'package:hive_ce/hive.dart';

/// Hive-backed local SoR for courier offline data.
///
/// Prefer Drift long-term; this store avoids build_runner for MVP analyze/compile.
class CourierLocalStore {
  CourierLocalStore._({
    required Box<dynamic> deliveries,
    required Box<dynamic> breadcrumbs,
    required Box<dynamic> syncQueue,
    required Box<dynamic> offerCache,
  })  : _deliveries = deliveries,
        _breadcrumbs = breadcrumbs,
        _syncQueue = syncQueue,
        _offerCache = offerCache;

  static const deliveriesBox = 'courier_active_deliveries';
  static const breadcrumbsBox = 'courier_location_breadcrumbs';
  static const syncQueueBox = 'courier_sync_queue';
  static const offerCacheBox = 'courier_offer_cache';

  final Box<dynamic> _deliveries;
  final Box<dynamic> _breadcrumbs;
  final Box<dynamic> _syncQueue;
  final Box<dynamic> _offerCache;

  static Future<CourierLocalStore> open({required String path}) async {
    Hive.init(path);
    final deliveries = await Hive.openBox<dynamic>(deliveriesBox);
    final breadcrumbs = await Hive.openBox<dynamic>(breadcrumbsBox);
    final syncQueue = await Hive.openBox<dynamic>(syncQueueBox);
    final offerCache = await Hive.openBox<dynamic>(offerCacheBox);
    return CourierLocalStore._(
      deliveries: deliveries,
      breadcrumbs: breadcrumbs,
      syncQueue: syncQueue,
      offerCache: offerCache,
    );
  }

  // --- Active deliveries ---

  Future<void> upsertDelivery(String id, Map<String, dynamic> snapshot) async {
    await _deliveries.put(id, snapshot);
  }

  Map<String, dynamic>? getDelivery(String id) {
    final raw = _deliveries.get(id);
    if (raw is Map) return Map<String, dynamic>.from(raw);
    return null;
  }

  List<Map<String, dynamic>> listDeliveries() {
    return _deliveries.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
  }

  Future<void> removeDelivery(String id) => _deliveries.delete(id);

  // --- Location breadcrumbs ---

  Future<void> enqueueBreadcrumb(Map<String, dynamic> point) async {
    final key = point['id']?.toString() ??
        DateTime.now().toUtc().microsecondsSinceEpoch.toString();
    await _breadcrumbs.put(key, point);
  }

  List<Map<String, dynamic>> peekBreadcrumbs({int limit = 100}) {
    final items = _breadcrumbs.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
    if (items.length <= limit) return items;
    return items.sublist(0, limit);
  }

  Future<void> clearBreadcrumbs(Iterable<String> ids) async {
    for (final id in ids) {
      await _breadcrumbs.delete(id);
    }
  }

  // --- Sync queue (local domain queue; mutations also use MutationOutbox) ---

  Future<void> enqueueSyncItem(String id, Map<String, dynamic> item) async {
    await _syncQueue.put(id, item);
  }

  List<Map<String, dynamic>> peekSyncQueue({int limit = 50}) {
    final items = _syncQueue.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
    if (items.length <= limit) return items;
    return items.sublist(0, limit);
  }

  Future<void> removeSyncItem(String id) => _syncQueue.delete(id);

  // --- Offer cache ---

  Future<void> cacheOffer(String id, Map<String, dynamic> offer) async {
    await _offerCache.put(id, offer);
  }

  Map<String, dynamic>? getCachedOffer(String id) {
    final raw = _offerCache.get(id);
    if (raw is Map) return Map<String, dynamic>.from(raw);
    if (raw is String) {
      try {
        return Map<String, dynamic>.from(jsonDecode(raw) as Map);
      } catch (_) {
        return null;
      }
    }
    return null;
  }

  List<Map<String, dynamic>> listCachedOffers() {
    return _offerCache.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
  }

  Future<void> removeCachedOffer(String id) => _offerCache.delete(id);

  Future<void> close() async {
    await _deliveries.close();
    await _breadcrumbs.close();
    await _syncQueue.close();
    await _offerCache.close();
  }
}
