import 'dart:convert';

import 'package:hive_ce/hive.dart';

/// Hive-backed local SoR for warehouse offline data.
///
/// Prefer Drift long-term; this store avoids build_runner for MVP analyze/compile.
class WarehouseLocalStore {
  WarehouseLocalStore._({
    required Box<dynamic> activePickTasks,
    required Box<dynamic> scanOutboxCache,
    required Box<dynamic> stationPrefs,
  })  : _activePickTasks = activePickTasks,
        _scanOutboxCache = scanOutboxCache,
        _stationPrefs = stationPrefs;

  static const activePickTasksBox = 'warehouse_active_pick_tasks';
  static const scanOutboxCacheBox = 'warehouse_scan_outbox_cache';
  static const stationPrefsBox = 'warehouse_station_prefs';

  final Box<dynamic> _activePickTasks;
  final Box<dynamic> _scanOutboxCache;
  final Box<dynamic> _stationPrefs;

  static Future<WarehouseLocalStore> open({required String path}) async {
    Hive.init(path);
    final activePickTasks = await Hive.openBox<dynamic>(activePickTasksBox);
    final scanOutboxCache = await Hive.openBox<dynamic>(scanOutboxCacheBox);
    final stationPrefs = await Hive.openBox<dynamic>(stationPrefsBox);
    return WarehouseLocalStore._(
      activePickTasks: activePickTasks,
      scanOutboxCache: scanOutboxCache,
      stationPrefs: stationPrefs,
    );
  }

  // --- Active pick tasks ---

  Future<void> upsertPickTask(String id, Map<String, dynamic> snapshot) async {
    await _activePickTasks.put(id, snapshot);
  }

  Map<String, dynamic>? getPickTask(String id) {
    final raw = _activePickTasks.get(id);
    if (raw is Map) return Map<String, dynamic>.from(raw);
    return null;
  }

  List<Map<String, dynamic>> listPickTasks() {
    return _activePickTasks.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
  }

  Future<void> removePickTask(String id) => _activePickTasks.delete(id);

  Future<void> clearPickTasks() => _activePickTasks.clear();

  // --- Scan outbox cache (optimistic scans pending SyncEngine flush) ---

  Future<void> enqueueScanEvent(String id, Map<String, dynamic> event) async {
    await _scanOutboxCache.put(id, event);
  }

  List<Map<String, dynamic>> peekScanEvents({int limit = 100}) {
    final items = _scanOutboxCache.values
        .whereType<Map>()
        .map((e) => Map<String, dynamic>.from(e))
        .toList();
    if (items.length <= limit) return items;
    return items.sublist(0, limit);
  }

  Future<void> removeScanEvent(String id) => _scanOutboxCache.delete(id);

  Future<void> clearScanEvents(Iterable<String> ids) async {
    for (final id in ids) {
      await _scanOutboxCache.delete(id);
    }
  }

  // --- Station prefs (last station, zone defaults, scan gun mode) ---

  Future<void> setStationPref(String key, Object? value) async {
    if (value == null) {
      await _stationPrefs.delete(key);
      return;
    }
    await _stationPrefs.put(key, value);
  }

  T? getStationPref<T>(String key, {T? defaultValue}) {
    final raw = _stationPrefs.get(key, defaultValue: defaultValue);
    if (raw is T) return raw;
    if (raw is Map && T == Map<String, dynamic>) {
      return Map<String, dynamic>.from(raw) as T;
    }
    if (raw is String && T == Map<String, dynamic>) {
      try {
        return Map<String, dynamic>.from(jsonDecode(raw) as Map) as T;
      } catch (_) {
        return defaultValue;
      }
    }
    return defaultValue;
  }

  String? get lastStationId => getStationPref<String>('last_station_id');

  Future<void> setLastStationId(String? stationId) =>
      setStationPref('last_station_id', stationId);

  Future<void> close() async {
    await _activePickTasks.close();
    await _scanOutboxCache.close();
    await _stationPrefs.close();
  }
}
