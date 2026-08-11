/// Resolves remote module manifests into lazy/hot-updatable entry points.
class ModuleLoader {
  const ModuleLoader();

  /// Loads a module by [entryPoint] scheme: flutter:// | wasm:// | mf://
  Future<void> load(String entryPoint, {required String version}) async {
    // Production: verify signature/checksum via superapp-service resolve payload.
    if (entryPoint.isEmpty) {
      throw ArgumentError('entryPoint required');
    }
  }

  Future<void> unload(String key) async {}
}
