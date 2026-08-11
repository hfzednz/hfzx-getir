import 'package:hive_ce/hive.dart';

/// Hive CE backed preferences and lightweight KV cache (CONSTITUTION §24).
class PreferencesStore {
  PreferencesStore._(this._box);

  static const defaultBoxName = 'nexora_preferences';

  final Box<dynamic> _box;

  static Future<PreferencesStore> open({
    required String path,
    String boxName = defaultBoxName,
  }) async {
    Hive.init(path);
    final box = await Hive.openBox<dynamic>(boxName);
    return PreferencesStore._(box);
  }

  T? get<T>(String key, {T? defaultValue}) {
    final value = _box.get(key, defaultValue: defaultValue);
    return value as T?;
  }

  Future<void> set<T>(String key, T value) => _box.put(key, value);

  Future<void> remove(String key) => _box.delete(key);

  Future<void> clear() => _box.clear();

  bool containsKey(String key) => _box.containsKey(key);

  Iterable<String> get keys => _box.keys.cast<String>();

  Future<void> close() => _box.close();
}
