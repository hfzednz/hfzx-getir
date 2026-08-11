/// Plugin / mini-app manifest contract mirrored from superapp-service.
class PluginManifest {
  const PluginManifest({
    required this.key,
    required this.version,
    required this.entryPoint,
    this.permissions = const [],
    this.hooks = const [],
    this.minShellVersion = '1.0.0',
  });

  final String key;
  final String version;
  final String entryPoint;
  final List<String> permissions;
  final List<String> hooks;
  final String minShellVersion;

  Map<String, Object?> toJson() => {
        'key': key,
        'version': version,
        'entryPoint': entryPoint,
        'permissions': permissions,
        'hooks': hooks,
        'minShellVersion': minShellVersion,
      };
}
