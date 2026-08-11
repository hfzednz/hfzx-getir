/// Extension hook names aligned with platform Extension API.
abstract final class ExtensionHooks {
  static const navigation = 'navigation';
  static const payment = 'payment';
  static const notification = 'notification';
  static const search = 'search';
  static const ai = 'ai';
  static const analytics = 'analytics';
  static const order = 'order';
  static const profile = 'profile';
}

typedef HookHandler = Future<void> Function(Map<String, Object?> payload);

class HookRegistry {
  final Map<String, HookHandler> _handlers = {};

  void register(String hook, HookHandler handler) => _handlers[hook] = handler;

  Future<void> invoke(String hook, Map<String, Object?> payload) async {
    final h = _handlers[hook];
    if (h != null) await h(payload);
  }
}
