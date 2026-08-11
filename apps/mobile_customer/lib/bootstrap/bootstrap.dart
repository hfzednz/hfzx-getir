import 'dart:async';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_crashlytics/firebase_crashlytics.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nexora_core/nexora_core.dart';
import 'package:path_provider/path_provider.dart';

import '../app/nexora_app.dart';
import '../di/providers.dart';
import '../features/auth/presentation/providers/auth_session_provider.dart';
import '../features/cart/data/local/app_database.dart';
import 'deep_link_handler.dart';
import 'fcm_bootstrap.dart';

class BootstrapResult {
  BootstrapResult({required this.overrides});
  final List<Override> overrides;
}

Future<BootstrapResult> bootstrap() async {
  WidgetsFlutterBinding.ensureInitialized();

  try {
    await dotenv.load(fileName: '.env.example');
  } catch (_) {
    // dotenv optional in dev when using dart-define
  }

  final appDir = await getApplicationDocumentsDirectory();
  final prefs = await PreferencesStore.open(path: appDir.path);
  final outbox = await HiveMutationOutbox.open(path: appDir.path);
  final database = AppDatabase();

  await _initFirebase();

  final overrides = <Override>[
    preferencesStoreProvider.overrideWithValue(prefs),
    mutationOutboxProvider.overrideWithValue(outbox),
    databaseProvider.overrideWithValue(database),
  ];

  return BootstrapResult(overrides: overrides);
}

Future<void> _initFirebase() async {
  try {
    if (Firebase.apps.isEmpty) {
      await Firebase.initializeApp();
    }
    if (!kIsWeb) {
      FlutterError.onError = FirebaseCrashlytics.instance.recordFlutterFatalError;
      PlatformDispatcher.instance.onError = (error, stack) {
        unawaited(
          FirebaseCrashlytics.instance.recordError(error, stack, fatal: true),
        );
        return true;
      };
    }
  } catch (_) {
    // Firebase options missing — skip messaging/crashlytics gracefully
  }
}

Future<void> runNexoraApp(BootstrapResult bootstrap) async {
  runApp(
    ProviderScope(
      overrides: bootstrap.overrides,
      child: const NexoraApp(),
    ),
  );
}

Future<void> postBootstrap(WidgetRef ref) async {
  await ref.read(authSessionProvider.notifier).restore();
  unawaited(ref.read(syncEngineProvider).flush());
  unawaited(registerFcmIfAvailable(ref));
  unawaited(ref.read(deepLinkHandlerProvider).start());
}
