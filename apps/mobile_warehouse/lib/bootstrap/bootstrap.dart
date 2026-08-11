import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:path_provider/path_provider.dart';

import '../app/nexora_warehouse_app.dart';
import '../data/local/warehouse_local_store.dart';
import '../di/providers.dart';

class BootstrapResult {
  BootstrapResult({required this.overrides});
  final List<Override> overrides;
}

Future<BootstrapResult> bootstrap() async {
  WidgetsFlutterBinding.ensureInitialized();

  try {
    await dotenv.load(fileName: '.env.example');
  } catch (_) {
    // Optional when using --dart-define.
  }

  final appDir = await getApplicationDocumentsDirectory();
  final prefs = await PreferencesStore.open(path: appDir.path);
  final outbox = await HiveMutationOutbox.open(path: appDir.path);
  final localStore = await WarehouseLocalStore.open(path: appDir.path);

  if (kDebugMode) {
    debugPrint('NEXORA Warehouse bootstrap complete');
  }

  return BootstrapResult(
    overrides: [
      preferencesStoreProvider.overrideWithValue(prefs),
      mutationOutboxProvider.overrideWithValue(outbox),
      warehouseLocalStoreProvider.overrideWithValue(localStore),
    ],
  );
}

Future<void> runNexoraWarehouseApp(BootstrapResult bootstrap) async {
  runApp(
    ProviderScope(
      overrides: bootstrap.overrides,
      child: const NexoraWarehouseApp(),
    ),
  );
}
